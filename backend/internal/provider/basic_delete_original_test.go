//go:build unit

package provider_test

import (
	"context"
	"errors"
	"testing"

	providercore "github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/stretchr/testify/require"
)

// providerRepoStub 是 ProviderRepository 接口的测试桩实现。
// 用于隔离测试 ProviderService.Delete 方法，避免依赖真实数据库。
//
// 设计说明：
//   - exists: 模拟 ExistsByID 返回的存在性结果
//   - existsErr: 模拟 ExistsByID 返回的错误
//   - deleteErr: 模拟 Delete 返回的错误
//   - deletedIDs: 记录被调用删除的提供商 ID，用于断言验证
type providerRepoStub struct {
	providercore.BasicProviderStore
	exists     bool    // ExistsByID 的返回值
	existsErr  error   // ExistsByID 的错误返回值
	deleteErr  error   // Delete 的错误返回值
	deletedIDs []int64 // 记录已删除的提供商 ID 列表
}

// TestProviderService_Delete_NotFound 测试删除不存在的提供商时返回正确的错误。
// 预期行为：
//   - ExistsByID 返回 false（提供商不存在）
//   - 返回 ErrProviderNotFound 错误
//   - Delete 方法不被调用（deletedIDs 为空）
func TestProviderService_Delete_NotFound(t *testing.T) {
	repo := &providerRepoStub{exists: false}
	svc := providercore.NewBasicProviders(repo, nil, nil)

	err := svc.Delete(context.Background(), 55)
	require.ErrorIs(t, err, providercore.ErrProviderNotFound)
	require.Empty(t, repo.deletedIDs)
}

// TestProviderService_Delete_CheckError 测试存在性检查失败时的错误处理。
// 预期行为：
//   - ExistsByID 返回数据库错误
//   - 返回包含 "check provider" 的错误信息
//   - Delete 方法不被调用
func TestProviderService_Delete_CheckError(t *testing.T) {
	repo := &providerRepoStub{existsErr: errors.New("db down")}
	svc := providercore.NewBasicProviders(repo, nil, nil)

	err := svc.Delete(context.Background(), 55)
	require.Error(t, err)
	require.ErrorContains(t, err, "check provider")
	require.Empty(t, repo.deletedIDs)
}

// TestProviderService_Delete_DeleteError 测试删除操作失败时的错误处理。
// 预期行为：
//   - ExistsByID 返回 true（提供商存在）
//   - Delete 被调用但返回错误
//   - 返回包含 "delete provider" 的错误信息
//   - deletedIDs 记录了尝试删除的 ID
func TestProviderService_Delete_DeleteError(t *testing.T) {
	repo := &providerRepoStub{
		exists:    true,
		deleteErr: errors.New("delete failed"),
	}
	svc := providercore.NewBasicProviders(repo, nil, nil)

	err := svc.Delete(context.Background(), 55)
	require.Error(t, err)
	require.ErrorContains(t, err, "delete provider")
	require.Equal(t, []int64{55}, repo.deletedIDs)
}

// TestProviderService_Delete_Success 测试删除操作成功的场景。
// 预期行为：
//   - ExistsByID 返回 true（提供商存在）
//   - Delete 成功执行
//   - 返回 nil 错误
//   - deletedIDs 记录了被删除的 ID
func TestProviderService_Delete_Success(t *testing.T) {
	repo := &providerRepoStub{exists: true}
	svc := providercore.NewBasicProviders(repo, nil, nil)

	err := svc.Delete(context.Background(), 55)
	require.NoError(t, err)
	require.Equal(t, []int64{55}, repo.deletedIDs)
}

// ExistsByID 保留原删除前存在性检查替身。
func (s *providerRepoStub) ExistsByID(context.Context, int64) (bool, error) {
	return s.exists, s.existsErr
}

// Delete 记录原删除调用顺序及结果。
func (s *providerRepoStub) Delete(_ context.Context, id int64) error {
	s.deletedIDs = append(s.deletedIDs, id)
	return s.deleteErr
}
