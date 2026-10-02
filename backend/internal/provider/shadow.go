package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	infraerrors "github.com/TokenFlux/TokenRouter/internal/pkg/apperror"
)

// CreateShadow 为指定 OpenAI OAuth 母提供商创建 spark 维度影子提供商（一母一影）。
// 安全不变量：Credentials 恒不含 auth token（仅 model_mapping，守卫 isAllowedSparkShadowCredentialsUpdate 放行）。
func (s *Admin) CreateShadow(ctx context.Context, parentID int64, opts ShadowOptions) (*Record, error) {
	// 1. 加载母提供商并校验平台/类型
	parent, err := s.providerRepo.GetByID(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("get parent provider: %w", err)
	}
	if !parent.IsOpenAIOAuth() {
		return nil, infraerrors.New(infraerrors.CategoryBadRequest, "SPARK_SHADOW_INVALID_PARENT",
			"spark shadow requires an OpenAI OAuth parent provider")
	}
	// 母提供商需要持有独立凭据，resolveCredentialProvider 解析一层母提供商。
	if parent.IsCredentialShadow() {
		return nil, infraerrors.New(infraerrors.CategoryBadRequest, "SPARK_SHADOW_PARENT_IS_SHADOW",
			"spark shadow parent must be a real provider, not another spark shadow")
	}

	// 2. 一母一影校验
	shadows, err := s.providerRepo.ListShadowsByParent(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("check existing spark shadows: %w", err)
	}
	if len(shadows) > 0 {
		return nil, infraerrors.New(infraerrors.CategoryConflict, "SPARK_SHADOW_ALREADY_EXISTS",
			"parent provider already has a spark shadow provider")
	}

	// 请求提供分组时先校验，省略时继承母提供商已有的分组。
	// 母提供商没有分组时保持未分组，不按名称寻找其他组。
	groupIDs := opts.GroupIDs
	if len(groupIDs) > 0 {
		if s.options.Groups != nil {
			if err := s.ValidateGroupIDs(ctx, groupIDs); err != nil {
				return nil, err
			}
		}
	} else if len(parent.GroupIDs) > 0 {
		groupIDs = append([]int64(nil), parent.GroupIDs...)
	}

	// 4. 构造影子提供商。Credentials 仅保存 model_mapping，不保存认证令牌。
	// 名称为空时使用“<母提供商名> (Spark)”，并按 rune 截取前 100 个字符。
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		name = parent.Name + " (Spark)"
	}
	if runes := []rune(name); len(runes) > 100 {
		name = string(runes[:100])
	}
	// 并发值小于等于零时继承母提供商的值，限流器将零解释为无限并发。
	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = parent.Concurrency
	}
	// 未指定优先级（<=0）时继承母提供商。前端只传名称创建时，Priority 的零值表示省略。
	// 较小的调度优先级数值优先，SetPriority 会写入传入值并覆盖数据库默认值 50，因此使用母提供商的优先级。
	// 代理始终继承母提供商，省略的分组和并发参数也沿用母提供商的值。
	priority := opts.Priority
	if priority <= 0 {
		priority = parent.Priority
	}
	shadow := &Record{
		Name:             name,
		Platform:         PlatformOpenAI,
		Type:             ProviderTypeOAuth,
		Status:           StatusActive,
		Credentials:      map[string]any{"model_mapping": s.options.ShadowModels()},
		ParentProviderID: &parentID,
		QuotaDimension:   QuotaDimensionSpark,
		ProxyID:          parent.ProxyID,
		Priority:         priority,
		Concurrency:      concurrency,
		Schedulable:      true,
	}

	// 5. 持久化并填充 shadow.ID。预查后若有并发请求抢先创建，唯一索引会拒绝本次写入。
	// 复查确认影子已经存在时返回结构化 409。
	if err := s.providerRepo.Create(ctx, shadow); err != nil {
		if existing, qerr := s.providerRepo.ListShadowsByParent(ctx, parentID); qerr == nil && len(existing) > 0 {
			return nil, infraerrors.New(infraerrors.CategoryConflict, "SPARK_SHADOW_ALREADY_EXISTS",
				"parent provider already has a spark shadow provider")
		}
		return nil, fmt.Errorf("create spark shadow: %w", err)
	}

	// 6. 绑定分组。创建和绑定分两次提交，绑定失败时尝试删除刚创建的影子，释放唯一索引后可重试。
	// 补偿删除使用独立取消上下文，请求取消或超时后仍会执行；进程崩溃时仍可能留下未完成的记录。
	if len(groupIDs) > 0 {
		if err := s.providerRepo.BindGroups(ctx, shadow.ID, groupIDs); err != nil {
			if delErr := s.providerRepo.Delete(context.WithoutCancel(ctx), shadow.ID); delErr != nil {
				s.options.Error("spark_shadow_bind_groups_rollback_failed",
					"shadow_id", shadow.ID, "parent_id", parentID, "delete_err", delErr)
			}
			return nil, fmt.Errorf("bind groups for spark shadow: %w", err)
		}
		shadow.GroupIDs = groupIDs
	}

	return shadow, nil
}

// propagateProxyToShadows syncs proxyID to all spark shadow providers of parentID.
// It is called synchronously so that proxy changes are immediately consistent;
// providerRepo.Update triggers the scheduler outbox + cache propagation internally.
// Calling this for a non-parent provider is a harmless no-op.
func (s *Admin) propagateProxyToShadows(ctx context.Context, parentID int64, proxyID *int64) error {
	return PropagateProviderProxyToShadows(ctx, s.providerRepo, parentID, proxyID)
}

// PropagateProviderProxyToShadows 将母提供商的代理同步到其 Spark 影子。
// 管理编辑和 CRS 同步都调用此入口，影子的出站代理随母提供商更新。
func PropagateProviderProxyToShadows(ctx context.Context, repo ShadowProxyStore, parentID int64, proxyID *int64) error {
	shadows, err := repo.ListShadowsByParent(ctx, parentID)
	if err != nil {
		return fmt.Errorf("list spark shadows for proxy propagation: %w", err)
	}
	for _, shadow := range shadows {
		shadow.ProxyID = proxyID
		if err := WriteConfiguration(ctx, repo, shadow, ConfigurationChange{Fields: ConfigProxyID}); err != nil {
			return fmt.Errorf("update spark shadow %d proxy: %w", shadow.ID, err)
		}
	}
	return nil
}

func (s *Admin) RevertProviderProxyFallback(ctx context.Context, id int64) error {
	if err := s.providerRepo.RevertProxyFallback(ctx, id); err != nil {
		return err
	}
	// 加载回退后的提供商以获取实际 ProxyID，再传播到影子提供商
	provider, err := s.providerRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get provider after proxy revert: %w", err)
	}
	return s.propagateProxyToShadows(ctx, id, provider.ProxyID)
}

func (s *Admin) ValidateGroupIDs(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	if s.options.Groups == nil {
		return errors.New("group repository not configured")
	}
	return s.options.Groups.ValidateGroups(ctx, ids)
}
