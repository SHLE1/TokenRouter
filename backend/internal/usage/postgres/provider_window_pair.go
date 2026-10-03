package postgres

import (
	"context"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/usage"
)

// GetProviderWindowStatsPair 扫描两个窗口的并集，分别计算各自起点后的统计。
func (r *Store) GetProviderWindowStatsPair(ctx context.Context, id int64, firstStart, secondStart time.Time) (*usage.ProviderStats, *usage.ProviderStats, error) {
	const query = `SELECT
	 COUNT(*) FILTER (WHERE created_at >= $2),
	 COALESCE(SUM(input_tokens+output_tokens+cache_creation_tokens+cache_read_tokens) FILTER (WHERE created_at >= $2),0),
	 COALESCE(SUM(COALESCE(provider_stats_cost,total_cost)*COALESCE(provider_rate_multiplier,1)) FILTER (WHERE created_at >= $2),0),
	 COALESCE(SUM(total_cost) FILTER (WHERE created_at >= $2),0),
	 COALESCE(SUM(actual_cost) FILTER (WHERE created_at >= $2),0),
	 COUNT(*) FILTER (WHERE created_at >= $3),
	 COALESCE(SUM(input_tokens+output_tokens+cache_creation_tokens+cache_read_tokens) FILTER (WHERE created_at >= $3),0),
	 COALESCE(SUM(COALESCE(provider_stats_cost,total_cost)*COALESCE(provider_rate_multiplier,1)) FILTER (WHERE created_at >= $3),0),
	 COALESCE(SUM(total_cost) FILTER (WHERE created_at >= $3),0),
	 COALESCE(SUM(actual_cost) FILTER (WHERE created_at >= $3),0)
	 FROM usage_logs WHERE provider_id=$1 AND created_at >= LEAST($2,$3)`
	a, b := &usage.ProviderStats{}, &usage.ProviderStats{}
	if err := scanSingleRow(ctx, r.sql, query, []any{id, firstStart, secondStart},
		&a.Requests, &a.Tokens, &a.Cost, &a.StandardCost, &a.UserCost,
		&b.Requests, &b.Tokens, &b.Cost, &b.StandardCost, &b.UserCost); err != nil {
		return nil, nil, err
	}
	return a, b, nil
}
