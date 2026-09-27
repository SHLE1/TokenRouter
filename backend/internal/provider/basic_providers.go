// 本文件维护 provider 的所属能力；兼容入口复用唯一实现。
package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/pkg/pagination"
)

// CreateProviderRequest 创建提供商请求
type CreateProviderRequest struct {
	Name               string         `json:"name"`
	Notes              *string        `json:"notes"`
	Platform           string         `json:"platform"`
	Type               string         `json:"type"`
	Credentials        map[string]any `json:"credentials"`
	Extra              map[string]any `json:"extra"`
	ProxyID            *int64         `json:"proxy_id"`
	Concurrency        int            `json:"concurrency"`
	Priority           int            `json:"priority"`
	GroupIDs           []int64        `json:"group_ids"`
	ExpiresAt          *time.Time     `json:"expires_at"`
	AutoPauseOnExpired *bool          `json:"auto_pause_on_expired"`
}

// UpdateProviderRequest 更新提供商请求
type UpdateProviderRequest struct {
	Name               *string         `json:"name"`
	Notes              *string         `json:"notes"`
	Credentials        *map[string]any `json:"credentials"`
	Extra              *map[string]any `json:"extra"`
	ProxyID            *int64          `json:"proxy_id"`
	Concurrency        *int            `json:"concurrency"`
	Priority           *int            `json:"priority"`
	Status             *string         `json:"status"`
	GroupIDs           *[]int64        `json:"group_ids"`
	ExpiresAt          *time.Time      `json:"expires_at"`
	AutoPauseOnExpired *bool           `json:"auto_pause_on_expired"`
}

// Create 创建提供商
func (s *BasicProviders) Create(ctx context.Context, req CreateProviderRequest) (*Record, error) {
	// 验证分组是否存在（如果指定了分组）
	if len(req.GroupIDs) > 0 {
		if err := s.validateGroupIDsExist(ctx, req.GroupIDs); err != nil {
			return nil, err
		}
	}

	// 创建提供商
	provider := &Record{
		Name:        req.Name,
		Notes:       NormalizeProviderNotes(req.Notes),
		Platform:    req.Platform,
		Type:        req.Type,
		Credentials: SanitizeStoredCredentials(req.Platform, req.Credentials),
		Extra:       PrepareCodexFingerprintExtraForCreate(req.Platform, req.Type, req.Extra, s.newSeed),
		ProxyID:     req.ProxyID,
		Concurrency: req.Concurrency,
		Priority:    req.Priority,
		Status:      StatusActive,
		ExpiresAt:   req.ExpiresAt,
	}
	if req.AutoPauseOnExpired != nil {
		provider.AutoPauseOnExpired = *req.AutoPauseOnExpired
	} else {
		provider.AutoPauseOnExpired = true
	}
	if err := NormalizeUpstreamUsageExtra(provider.Extra); err != nil {
		return nil, err
	}

	if err := NormalizeCNProviderCredentials(provider, true); err != nil {
		return nil, err
	}
	if err := NormalizeOpenAIAPIKeyConfiguration(provider); err != nil {
		return nil, err
	}
	if err := NormalizeProviderProtocols(provider); err != nil {
		return nil, err
	}
	if err := s.providerRepo.Create(ctx, provider); err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	// require_oauth_only 检查：apikey 类型提供商不可加入限制分组
	if provider.Type == ProviderTypeAPIKey && len(req.GroupIDs) > 0 {
		for _, gid := range req.GroupIDs {
			g, err := s.groupRepo.GetGroup(ctx, gid)
			if err != nil {
				return nil, err
			}
			if g.RequireOAuthOnly {
				return nil, fmt.Errorf("分组 [%s] 仅允许 OAuth 提供商，apikey 类型提供商无法加入", g.Name)
			}
		}
	}

	// 绑定分组
	if len(req.GroupIDs) > 0 {
		if err := s.providerRepo.BindGroups(ctx, provider.ID, req.GroupIDs); err != nil {
			return nil, fmt.Errorf("bind groups: %w", err)
		}
	}

	return provider, nil
}

// GetByID 根据ID获取提供商
func (s *BasicProviders) GetByID(ctx context.Context, id int64) (*Record, error) {
	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get provider: %w", err)
	}
	return provider, nil
}

// List 获取提供商列表
func (s *BasicProviders) List(ctx context.Context, params pagination.PaginationParams) ([]Record, *pagination.PaginationResult, error) {
	providers, pagination, err := s.providerRepo.List(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("list providers: %w", err)
	}
	return providers, pagination, nil
}

// ListByPlatform 根据平台获取提供商列表
func (s *BasicProviders) ListByPlatform(ctx context.Context, platform string) ([]Record, error) {
	providers, err := s.providerRepo.ListByPlatform(ctx, platform)
	if err != nil {
		return nil, fmt.Errorf("list providers by platform: %w", err)
	}
	return providers, nil
}

// ListByGroup 根据分组获取提供商列表
func (s *BasicProviders) ListByGroup(ctx context.Context, groupID int64) ([]Record, error) {
	providers, err := s.providerRepo.ListByGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("list providers by group: %w", err)
	}
	return providers, nil
}

// Update 更新提供商
func (s *BasicProviders) Update(ctx context.Context, id int64, req UpdateProviderRequest) (*Record, error) {
	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get provider: %w", err)
	}

	// 更新字段
	if req.Name != nil {
		provider.Name = *req.Name
	}
	if req.Notes != nil {
		provider.Notes = NormalizeProviderNotes(req.Notes)
	}

	if req.Credentials != nil {
		provider.Credentials = SanitizeStoredCredentials(provider.Platform, PreserveProtocolCredentials(provider.Credentials, *req.Credentials))
	}

	if req.Extra != nil {
		extra := make(map[string]any, len(*req.Extra))
		for key, value := range *req.Extra {
			extra[key] = value
		}
		delete(extra, OllamaCloudUsageSessionExtraKey)
		delete(extra, OllamaCloudUsageAutoRefreshExtraKey)
		delete(extra, OllamaCloudUsageSnapshotExtraKey)
		if err := NormalizeUpstreamUsageExtra(extra); err != nil {
			return nil, err
		}
		if _, provided := (*req.Extra)[UpstreamUsageQueryExtraKey]; !provided && provider.Extra != nil {
			if value, exists := provider.Extra[UpstreamUsageQueryExtraKey]; exists {
				if normalized, ok := NormalizedUpstreamUsageConfigValue(value); ok {
					extra[UpstreamUsageQueryExtraKey] = normalized
				}
			}
		}
		provider.Extra = PrepareCodexFingerprintExtraForUpdate(provider, extra, s.newSeed)
	} else {
		provider.Extra = PrepareCodexFingerprintExtraForUpdate(provider, provider.Extra, s.newSeed)
	}

	if req.ProxyID != nil {
		provider.ProxyID = req.ProxyID
	}

	if req.Concurrency != nil {
		provider.Concurrency = *req.Concurrency
	}

	if req.Priority != nil {
		provider.Priority = *req.Priority
	}

	if req.Status != nil {
		provider.Status = *req.Status
	}
	if req.ExpiresAt != nil {
		provider.ExpiresAt = req.ExpiresAt
	}
	if req.AutoPauseOnExpired != nil {
		provider.AutoPauseOnExpired = *req.AutoPauseOnExpired
	}

	// 先验证分组是否存在（在任何写操作之前）
	if req.GroupIDs != nil {
		if err := s.validateGroupIDsExist(ctx, *req.GroupIDs); err != nil {
			return nil, err
		}
	}

	if err := NormalizeCNProviderCredentials(provider, false); err != nil {
		return nil, err
	}
	if err := NormalizeOpenAIAPIKeyConfiguration(provider); err != nil {
		return nil, err
	}
	var credentialPatch, extraPatch map[string]any
	if req.Credentials != nil {
		credentialPatch = *req.Credentials
	}
	if req.Extra != nil {
		extraPatch = *req.Extra
	}
	ApplyLegacyProtocolPatch(provider, credentialPatch, extraPatch)
	if err := NormalizeProviderProtocols(provider); err != nil {
		return nil, err
	}
	// 只传递本次配置字段，避免完整对象回写并发凭据及健康状态。
	change := ConfigurationChange{NormalizeProtocols: true, CredentialInput: credentialPatch, ProtocolExtra: extraPatch}
	if req.Extra != nil {
		if _, provided := (*req.Extra)[UpstreamUsageQueryExtraKey]; !provided {
			change.PreserveExtraKeys = []string{UpstreamUsageQueryExtraKey}
		}
	}
	if req.Name != nil {
		change.Fields |= ConfigName
	}
	if req.Notes != nil {
		change.Fields |= ConfigNotes
	}
	if req.Credentials != nil {
		change.Fields |= ConfigCredentials
	}
	if req.Extra != nil {
		change.Fields |= ConfigExtra
	}
	if req.ProxyID != nil {
		change.Fields |= ConfigProxyID
	}
	if req.Concurrency != nil {
		change.Fields |= ConfigConcurrency
	}
	if req.Priority != nil {
		change.Fields |= ConfigPriority
	}
	if req.Status != nil {
		change.Fields |= ConfigStatus
	}
	if req.ExpiresAt != nil {
		change.Fields |= ConfigExpiresAt
	}
	if req.AutoPauseOnExpired != nil {
		change.Fields |= ConfigAutoPauseOnExpired
	}
	if err := WriteConfiguration(ctx, s.providerRepo, provider, change); err != nil {
		return nil, fmt.Errorf("update provider: %w", err)
	}

	// require_oauth_only 检查
	if provider.Type == ProviderTypeAPIKey && req.GroupIDs != nil {
		for _, gid := range *req.GroupIDs {
			g, err := s.groupRepo.GetGroup(ctx, gid)
			if err != nil {
				return nil, err
			}
			if g.RequireOAuthOnly {
				return nil, fmt.Errorf("分组 [%s] 仅允许 OAuth 提供商，apikey 类型提供商无法加入", g.Name)
			}
		}
	}

	// 绑定分组
	if req.GroupIDs != nil {
		if err := s.providerRepo.BindGroups(ctx, provider.ID, *req.GroupIDs); err != nil {
			return nil, fmt.Errorf("bind groups: %w", err)
		}
	}

	return provider, nil
}

// Delete 删除提供商
// 优化：使用 ExistsByID 替代 GetByID 进行存在性检查，
// 避免加载完整提供商对象及其关联数据，提升删除操作的性能
func (s *BasicProviders) Delete(ctx context.Context, id int64) error {
	// 使用轻量级的存在性检查，而非加载完整提供商对象
	exists, err := s.providerRepo.ExistsByID(ctx, id)
	if err != nil {
		return fmt.Errorf("check provider: %w", err)
	}
	// 明确返回提供商不存在错误，便于调用方区分错误类型
	if !exists {
		return ErrProviderNotFound
	}

	// 注意:此处不级联删除 spark 影子提供商。当前唯一的后台删除入口走 AdminService.DeleteProvider
	// (已 ListShadowsByParent 先删影子再删母)。本方法目前无删除调用方;若未来有调用方经此
	// 删除母提供商,需在此补级联,否则会留下孤儿影子(外审第6轮 P3:当前不可达,记为残留)。
	if err := s.providerRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete provider: %w", err)
	}

	return nil
}

func (s *BasicProviders) validateGroupIDsExist(ctx context.Context, groupIDs []int64) error {
	return s.groupRepo.ValidateGroups(ctx, groupIDs)
}

// UpdateStatus 更新提供商状态
func (s *BasicProviders) UpdateStatus(ctx context.Context, id int64, status string, errorMessage string) error {
	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get provider: %w", err)
	}

	provider.Status = status
	provider.ErrorMessage = errorMessage

	if err := WriteConfiguration(ctx, s.providerRepo, provider, ConfigurationChange{Fields: ConfigStatus | ConfigErrorMessage}); err != nil {
		return fmt.Errorf("update provider: %w", err)
	}

	return nil
}

// UpdateLastUsed 更新最后使用时间
func (s *BasicProviders) UpdateLastUsed(ctx context.Context, id int64) error {
	if err := s.providerRepo.UpdateLastUsed(ctx, id); err != nil {
		return fmt.Errorf("update last used: %w", err)
	}
	return nil
}

// GetCredential 获取提供商凭证（安全访问）
func (s *BasicProviders) GetCredential(ctx context.Context, id int64, key string) (string, error) {
	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("get provider: %w", err)
	}

	return provider.GetCredential(key), nil
}

// TestCredentials 测试提供商凭证是否有效（需要实现具体平台的测试逻辑）
func (s *BasicProviders) TestCredentials(ctx context.Context, id int64) error {
	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get provider: %w", err)
	}

	// 根据平台执行不同的测试逻辑
	switch provider.Platform {
	case PlatformAnthropic:
		// TODO: 测试Anthropic API凭证
		return nil
	case PlatformOpenAI:
		// TODO: 测试OpenAI API凭证
		return nil
	case PlatformGemini:
		// TODO: 测试Gemini API凭证
		return nil
	case PlatformGrok:
		// Grok OAuth 凭证通过 token 兑换、刷新和请求路径探测校验。
		return nil
	case PlatformKimi, PlatformZhipu, PlatformDeepseek:
		// 国产 OpenAI 兼容供应商：凭证为 API Key，实际可用性经余额/额度探测与转发路径验证。
		return nil
	default:
		return fmt.Errorf("unsupported platform: %s", provider.Platform)
	}
}

// BasicProviders 保留旧基础入口的查询顺序和错误包装，无缓存或后台状态。
// 生产管理链使用 Admin；此接口提供基础提供商操作。
type BasicProviders struct {
	providerRepo BasicProviderStore
	groupRepo    BasicProviderGroups
	newSeed      func() string
}
type BasicProviderGroups interface {
	GetGroup(context.Context, int64) (*GroupReference, error)
	ValidateGroups(context.Context, []int64) error
}
type BasicProviderStore interface {
	Create(context.Context, *Record) error
	GetByID(context.Context, int64) (*Record, error)
	List(context.Context, pagination.PaginationParams) ([]Record, *pagination.PaginationResult, error)
	ListByPlatform(context.Context, string) ([]Record, error)
	ListByGroup(context.Context, int64) ([]Record, error)
	Update(context.Context, *Record) error
	BindGroups(context.Context, int64, []int64) error
	ExistsByID(context.Context, int64) (bool, error)
	Delete(context.Context, int64) error
	UpdateLastUsed(context.Context, int64) error
}

func NewBasicProviders(store BasicProviderStore, groups BasicProviderGroups, newSeed func() string) *BasicProviders {
	return &BasicProviders{providerRepo: store, groupRepo: groups, newSeed: newSeed}
}
