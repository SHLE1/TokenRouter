package app

import (
	"context"
	"slices"
	"sync"

	"github.com/TokenFlux/TokenRouter/internal/gateway/provider"
)

// selectionGroupFixture 给旧的单平台仓储替身补齐批量平台查询与关系投影，选号仍执行生产策略。
// 源记录已有分组时原样保留，防止掩盖组外账号拒绝测试。
type selectionGroupFixture struct {
	provider.ExecutionAccountStore
	mu     sync.Mutex
	groups map[int64][]int64
}

func withSelectionGroupFixture(source provider.ExecutionAccountStore) provider.ExecutionAccountStore {
	if _, ok := source.(interface{ completeGroupProjection() }); ok {
		return source
	}
	if source == nil {
		return nil
	}
	return &selectionGroupFixture{ExecutionAccountStore: source, groups: make(map[int64][]int64)}
}

func (s *selectionGroupFixture) ListSchedulableByGroupIDAndPlatforms(ctx context.Context, id int64, platforms []string) ([]provider.ExecutionAccount, error) {
	values, err := s.ListSchedulableByGroupIDAndPlatform(ctx, id, "")
	if err != nil {
		return nil, err
	}
	var result []provider.ExecutionAccount
	for _, value := range values {
		if !slices.Contains(platforms, value.Record.Platform) {
			continue
		}
		copy := *provider.NewExecutionAccount(&value.Record)
		if len(copy.Record.GroupIDs) == 0 {
			copy.Record.GroupIDs = []int64{id}
			s.mu.Lock()
			s.groups[copy.Record.ID] = []int64{id}
			s.mu.Unlock()
		}
		result = append(result, copy)
	}
	return result, nil
}

func (s *selectionGroupFixture) ListSchedulableByGroupID(ctx context.Context, id int64) ([]provider.ExecutionAccount, error) {
	return s.ListSchedulableByGroupIDAndPlatforms(ctx, id, []string{"anthropic", "openai", "gemini", "antigravity", "qoder", "grok", "kimi", "zhipu", "deepseek"})
}

func (s *selectionGroupFixture) GetByID(ctx context.Context, id int64) (*provider.ExecutionAccount, error) {
	value, err := s.ExecutionAccountStore.GetByID(ctx, id)
	if value == nil || err != nil {
		return value, err
	}
	value = provider.NewExecutionAccount(&value.Record)
	if len(value.Record.GroupIDs) == 0 {
		s.mu.Lock()
		value.Record.GroupIDs = slices.Clone(s.groups[id])
		s.mu.Unlock()
	}
	return value, nil
}
