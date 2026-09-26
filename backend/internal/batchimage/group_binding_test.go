//go:build unit

package batchimage_test

import (
	"context"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/batchimage"
	"github.com/stretchr/testify/require"
)

// TestBatchImageUnboundKeyCannotUseGlobalAccounts 确保未绑定 Key 不借用全局账号池。
func TestBatchImageUnboundKeyCannotUseGlobalAccounts(t *testing.T) {
	svc, repo, _, provider, _, _ := newTestBatchImagePublicService(true)
	owner := testBatchImageOwner()
	owner.GroupID = nil
	_, err := svc.Submit(context.Background(), owner, validBatchImageSubmitRequest(), "")
	require.ErrorIs(t, err, batchimage.ErrBatchImageGroupDisabled)
	require.Empty(t, repo.jobs)
	require.Empty(t, provider.submits)
}
