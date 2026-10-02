package provider

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
	"github.com/stretchr/testify/require"
)

func TestAnthropicNegativeUsageCacheDoesNotCrossCredentialIdentity(t *testing.T) {
	oldErr := errors.New("old identity failed")
	cache := NewOAuthUsageCache()
	svc := NewOAuthUsageService(nil, cache, nil, OAuthUsageOptions{})
	cache.StoreAPI(int64(886), &OAuthAPIUsageCache{Identity: "old identity", Err: oldErr, Timestamp: time.Now()})
	a := &Record{ID: 886, Platform: capability.PlatformAnthropic, Type: capability.ProviderTypeOAuth, Status: StatusActive, Credentials: map[string]any{"access_token": "second"}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := svc.GetUsageForProvider(ctx, a, false)
	require.ErrorIs(t, err, context.Canceled)
}

// 查询期间管理员修改身份后，条件写入会拒绝此前身份的用量结果。
type activePassiveIdentityRepo struct {
	sessionWindowSyncRepo
	current Record
}

func (r *activePassiveIdentityRepo) GetByID(context.Context, int64) (*Record, error) {
	v := r.current
	return &v, nil
}

func TestAnthropicActiveUsageDoesNotWriteNewCredentialIdentity(t *testing.T) {
	a := Record{ID: 887, Platform: capability.PlatformAnthropic, Type: capability.ProviderTypeOAuth, Status: StatusActive, Credentials: map[string]any{"access_token": "old"}}
	repo := &activePassiveIdentityRepo{current: a}
	repo.current.Credentials = map[string]any{"access_token": "administrator"}
	cache := NewOAuthUsageCache()
	stats := NewLocalUsageStatistics(usageBatchStatisticsFixture{}, cache, LocalUsageStatisticsOptions{Now: time.Now, Log: log.Printf})
	svc := NewOAuthUsageService(repo, cache, stats, OAuthUsageOptions{})
	response := &ClaudeUsageResponse{}
	response.FiveHour.Utilization = 31
	response.FiveHour.ResetsAt = time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	cache.StoreAPI(a.ID, &OAuthAPIUsageCache{Identity: UsageCacheIdentity(&a), Response: response, Timestamp: time.Now()})
	_, err := svc.GetUsageForProvider(context.Background(), &a, false)
	require.NoError(t, err)
	require.Empty(t, repo.extraUpdates)
	require.Empty(t, repo.sessionWindowEnds)
}

// UpdateUsageExtraIfUnchanged 为查询测试模拟条件写入，SQL 行为由数据库测试覆盖。
func (r *activePassiveIdentityRepo) UpdateUsageExtraIfUnchanged(ctx context.Context, v UsageObservationVersion, updates map[string]any) (bool, error) {
	if !MatchesCredentialVersion(&r.current, v.CredentialVersion) {
		return false, nil
	}
	return true, r.UpdateExtra(ctx, v.ID, updates)
}

func (r *activePassiveIdentityRepo) UpdateUsageSessionWindowEndIfUnchanged(ctx context.Context, v UsageObservationVersion, _ *time.Time, end time.Time) (bool, error) {
	if !MatchesCredentialVersion(&r.current, v.CredentialVersion) {
		return false, nil
	}
	return true, r.UpdateSessionWindowEnd(ctx, v.ID, end)
}
