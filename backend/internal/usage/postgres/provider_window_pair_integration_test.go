//go:build integration

package postgres

import (
	"time"

	"github.com/TokenFlux/TokenRouter/internal/apikey"
	"github.com/TokenFlux/TokenRouter/internal/identity"
	"github.com/TokenFlux/TokenRouter/internal/provider"
)

// TestProviderWindowPairMatchesSingles 对照窗口起点、空窗口、反向起点、费用倍率和未来记录。
func (s *UsageLogRepoSuite) TestProviderWindowPairMatchesSingles() {
	u := mustCreateUser(s.T(), s.client, &identity.User{Email: "window-pair@test.local"})
	k := mustCreateApiKey(s.T(), s.client, &apikey.APIKey{UserID: u.ID, Key: "window-pair"})
	p := mustCreateProvider(s.T(), s.client, &provider.Record{Name: "window-pair"})
	empty := mustCreateProvider(s.T(), s.client, &provider.Record{Name: "empty-window-pair"})
	now := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	for i, at := range []time.Time{now.Add(-8 * 24 * time.Hour), now.Add(-7 * 24 * time.Hour), now.Add(-6 * 24 * time.Hour), now.Add(-5 * time.Hour), now.Add(-4 * time.Hour), now.Add(time.Hour)} {
		log := s.createUsageLog(u, k, p, 10+i, 20+i, float64(i), at)
		_, err := s.tx.ExecContext(s.ctx, "UPDATE usage_logs SET provider_stats_cost=2,provider_rate_multiplier=3,total_cost=4,cache_creation_tokens=5,cache_read_tokens=6 WHERE id=$1", log.ID)
		s.Require().NoError(err)
	}
	first, second := now.Add(-5*time.Hour), now.Add(-7*24*time.Hour)
	for _, id := range []int64{p.ID, empty.ID} {
		for _, starts := range [][2]time.Time{{first, second}, {second, first}, {first, first}, {now.Add(2 * time.Hour), now.Add(3 * time.Hour)}} {
			wantA, err := s.repo.GetProviderWindowStats(s.ctx, id, starts[0])
			s.Require().NoError(err)
			wantB, err := s.repo.GetProviderWindowStats(s.ctx, id, starts[1])
			s.Require().NoError(err)
			a, b, err := s.repo.GetProviderWindowStatsPair(s.ctx, id, starts[0], starts[1])
			s.Require().NoError(err)
			s.Require().Equal(wantA, a)
			s.Require().Equal(wantB, b)
		}
	}
}
