package httpapi

import (
	"context"
	"sync"

	"github.com/TokenFlux/TokenRouter/internal/provider"
)

// 创建测试记录提交给用例的输入，其他接口留空。
type managementCreateFixture struct {
	ProviderManagement
	mu                sync.Mutex
	createdProviders  []*provider.CreateProviderInput
	createProviderErr error
}

func (s *managementCreateFixture) CreateProvider(_ context.Context, input *provider.CreateProviderInput) (*provider.Record, error) {
	s.mu.Lock()
	s.createdProviders = append(s.createdProviders, input)
	s.mu.Unlock()
	if s.createProviderErr != nil {
		return nil, s.createProviderErr
	}
	return &provider.Record{ID: 300, Name: input.Name, Status: provider.StatusActive}, nil
}

func (s *managementCreateFixture) ForceOpenAIPrivacy(context.Context, *provider.Record) string {
	return ""
}

func (s *managementCreateFixture) ForceAntigravityPrivacy(context.Context, *provider.Record) string {
	return ""
}

// newManagementCreateFixtureHandler 组合管理 HTTP、批量操作和展示组件。
func newManagementCreateFixtureHandler(source *managementCreateFixture) *ManagementHandler {
	presenter := NewRuntimePresenter(provider.NewRuntimeStatusReader(provider.RuntimeStatusOptions{}), source, nil)
	batch := provider.NewManagementBatch(source, nil, provider.ManagementCreationOptions{Privacy: source})
	return NewManagementHandler(source, ManagementOptions{Presenter: presenter, Privacy: source, Batch: batch})
}
