//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/apikey"
	"github.com/TokenFlux/TokenRouter/internal/identity"
	"github.com/TokenFlux/TokenRouter/internal/pkg/timezone"
	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/settings/preaggregation"
	"github.com/TokenFlux/TokenRouter/internal/team"
	"github.com/TokenFlux/TokenRouter/internal/usage"
	"github.com/stretchr/testify/require"
)

// TestProviderReportRestoresConnectionJIT 在单连接池中确认报表事务结束后恢复连接设置。
func TestProviderReportRestoresConnectionJIT(t *testing.T) {
	limit := integrationDB.Stats().MaxOpenConnections
	integrationDB.SetMaxOpenConns(1)
	t.Cleanup(func() { integrationDB.SetMaxOpenConns(limit) })
	client := testEntClient(t)
	ctx := context.Background()
	p := mustCreateProvider(t, client, &provider.Record{Name: "report-jit"})
	var before, after string
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SHOW jit").Scan(&before))
	repo := NewUsageLogRepositoryWithSQL(client, integrationDB, timezone.NewCalendar(time.UTC))
	result, err := repo.GetProviderUsageStats(ctx, p.ID, time.Now().Add(-time.Hour), time.Now())
	require.NoError(t, err)
	require.Zero(t, result.Summary.TotalRequests)
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SHOW jit").Scan(&after))
	require.Equal(t, before, after)
}

// TestTeamReportsAnalyticsParity 对照原始结果检查团队、成员、Key、历史成员及时间桶两端。
func (s *UsageLogRepoSuite) TestTeamReportsAnalyticsParity() {
	owner := mustCreateUser(s.T(), s.client, &identity.User{Email: "report-owner@test.local"})
	member := mustCreateUser(s.T(), s.client, &identity.User{Email: "report-member@test.local"})
	former := mustCreateUser(s.T(), s.client, &identity.User{Email: "report-former@test.local"})
	empty := mustCreateUser(s.T(), s.client, &identity.User{Email: "report-empty@test.local"})
	entity, err := s.client.Team.Create().SetName("报表团队").SetStatus("active").SetMemberLimit(10).Save(s.ctx)
	s.Require().NoError(err)
	for _, u := range []*identity.User{owner, member, empty} {
		role := team.TeamRoleMember
		if u.ID == owner.ID {
			role = team.TeamRoleOwner
		}
		_, err = s.client.TeamMembership.Create().SetTeamID(entity.ID).SetUserID(u.ID).SetRole(role).Save(s.ctx)
		s.Require().NoError(err)
	}
	a := mustCreateApiKey(s.T(), s.client, &apikey.APIKey{UserID: owner.ID, TeamID: &entity.ID, Key: "report-a"})
	b := mustCreateApiKey(s.T(), s.client, &apikey.APIKey{UserID: owner.ID, TeamID: &entity.ID, Key: "report-b"})
	p := mustCreateProvider(s.T(), s.client, &provider.Record{Name: "report"})
	start := time.Date(2026, 9, 1, 10, 20, 0, 0, time.UTC)
	end := start.Add(96 * time.Hour)
	coverage := start.Add(12 * time.Hour).Truncate(time.Hour)
	watermark := end.Add(-2 * time.Hour)
	for i, actor := range []*identity.User{owner, member, former} {
		for j, at := range []time.Time{start.Add(-time.Second), start, start.Add(24 * time.Hour), watermark.Add(-time.Minute), watermark.Add(time.Minute), end} {
			keyID := a.ID
			if j%2 == 0 {
				keyID = b.ID
			}
			_, err = s.repo.Create(s.ctx, &usage.UsageLog{UserID: actor.ID, BillingUserID: owner.ID, TeamID: &entity.ID, APIKeyID: keyID, ProviderID: p.ID, RequestID: fmt.Sprintf("team-report-%d-%d", i, j), Model: "test", CreatedAt: at, InputTokens: 10, OutputTokens: 20, ActualCost: float64(j), TotalCost: float64(j)})
			s.Require().NoError(err)
		}
	}
	aggregation := NewAggregationStoreWithSQL(s.tx, s.repo.calendar)
	s.Require().NoError(aggregation.AggregateUsageAnalyticsRange(s.ctx, coverage, watermark))
	s.Require().NoError(aggregation.SaveUsageAnalyticsAggregationState(s.ctx, &usage.UsageAnalyticsAggregationState{CoverageStart: &coverage, LiveWatermark: watermark, Phase: "idle"}))
	settings := preaggregation.NewPreAggregationSettingsService(nil, &preaggregation.Options{Usage: preaggregation.UsageOptions{Enabled: true}})
	for _, zone := range []string{"UTC", "America/New_York", "Asia/Kathmandu"} {
		loc, err := time.LoadLocation(zone)
		s.Require().NoError(err)
		s.repo.calendar = timezone.NewCalendar(loc)
		for _, request := range []team.TeamUsageQuery{
			{From: start, To: end},
			{From: start, To: end, ActorUserID: &member.ID},
			{From: start, To: end, APIKeyID: &a.ID},
			{From: start, To: end, APIKeyID: &b.ID, ActorUserID: &member.ID},
		} {
			s.repo.preAggregation = nil
			want, err := s.repo.GetTeamUsageSummary(s.ctx, entity.ID, request)
			s.Require().NoError(err)
			members, err := s.repo.ListTeamMemberUsageSeries(s.ctx, entity.ID, request)
			s.Require().NoError(err)
			s.repo.preAggregation = settings
			_, ok, err := s.repo.teamReportSource(s.ctx, entity.ID, request)
			s.Require().NoError(err)
			s.Require().Equal(zone != "Asia/Kathmandu", ok)
			got, err := s.repo.GetTeamUsageSummary(s.ctx, entity.ID, request)
			s.Require().NoError(err)
			s.Require().Equal(want, got)
			gotMembers, err := s.repo.ListTeamMemberUsageSeries(s.ctx, entity.ID, request)
			s.Require().NoError(err)
			s.Require().Equal(members, gotMembers)
			if request.ActorUserID != nil {
				s.Require().Len(gotMembers, 1)
				s.Require().Equal(member.ID, gotMembers[0].ActorUserID)
			} else {
				s.Require().Len(gotMembers, 4)
				s.Require().Equal("left", gotMembers[2].Status)
				s.Require().Zero(gotMembers[3].Summary.RequestCount)
			}
		}
	}
	// 用量全部清除后，旧聚合桶不能产生历史成员或非零合计。
	s.repo.calendar = timezone.NewCalendar(time.UTC)
	_, err = s.tx.ExecContext(s.ctx, "DELETE FROM usage_logs WHERE team_id=$1", entity.ID)
	s.Require().NoError(err)
	request := team.TeamUsageQuery{From: start, To: end}
	got, err := s.repo.GetTeamUsageSummary(s.ctx, entity.ID, request)
	s.Require().NoError(err)
	s.Require().Zero(got.RequestCount)
	members, err := s.repo.ListTeamMemberUsageSeries(s.ctx, entity.ID, request)
	s.Require().NoError(err)
	s.Require().Len(members, 3)
	for _, item := range members {
		s.Require().Zero(item.Summary.RequestCount)
	}
}

// TestProviderReportCombinedResults 对照独立聚合，检查费用归属、空端点及非空耗时分母。
func (s *UsageLogRepoSuite) TestProviderReportCombinedResults() {
	u := mustCreateUser(s.T(), s.client, &identity.User{Email: "provider-report@test.local"})
	k := mustCreateApiKey(s.T(), s.client, &apikey.APIKey{UserID: u.ID, Key: "provider-report"})
	p := mustCreateProvider(s.T(), s.client, &provider.Record{Name: "provider-report"})
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(72 * time.Hour)
	for i := range 4 {
		log := s.createUsageLog(u, k, p, 10+i, 20+i, float64(i), start.Add(time.Duration(i)*12*time.Hour))
		_, err := s.tx.ExecContext(s.ctx, `UPDATE usage_logs SET model=$1,inbound_endpoint=$2,upstream_endpoint=$3,
			provider_stats_cost=2,provider_rate_multiplier=3,total_cost=4,cache_creation_tokens=5,cache_read_tokens=6,
			duration_ms=CASE WHEN $4=0 THEN NULL ELSE $4*100 END WHERE id=$5`, fmt.Sprintf("model-%d", i%2), []string{"", " /v1/messages "}[i%2], []string{" /v1/responses ", ""}[i%2], i, log.ID)
		s.Require().NoError(err)
	}
	wantModels, err := s.repo.GetModelStatsWithFilters(s.ctx, start, end, 0, 0, p.ID, 0, nil, nil, nil)
	s.Require().NoError(err)
	wantInbound, err := s.repo.GetEndpointStatsWithFilters(s.ctx, start, end, 0, 0, p.ID, 0, "", nil, nil, nil)
	s.Require().NoError(err)
	wantUpstream, err := s.repo.GetUpstreamEndpointStatsWithFilters(s.ctx, start, end, 0, 0, p.ID, 0, "", nil, nil, nil)
	s.Require().NoError(err)
	got, err := s.repo.GetProviderUsageStats(s.ctx, p.ID, start, end)
	s.Require().NoError(err)
	s.Require().ElementsMatch(wantModels, got.Models)
	s.Require().ElementsMatch(wantInbound, got.Endpoints)
	s.Require().ElementsMatch(wantUpstream, got.UpstreamEndpoints)
	s.Require().EqualValues(4, got.Summary.TotalRequests)
	s.Require().EqualValues(24, got.Summary.TotalCost)
	s.Require().EqualValues(6, got.Summary.TotalUserCost)
	s.Require().EqualValues(16, got.Summary.TotalStandardCost)
	s.Require().EqualValues(200, got.Summary.AvgDurationMs)
}
