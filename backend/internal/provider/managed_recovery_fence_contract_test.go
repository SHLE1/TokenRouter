package provider_test

import (
	"testing"
	"time"

	acctcore "github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
	"github.com/stretchr/testify/require"
)

// TestManagedRecoveryFencePreservesNewRuntimeBlock 检查恢复期间出现新阻断时该状态保持不变，并检查管理员清理操作。
func TestManagedRecoveryFencePreservesNewRuntimeBlock(t *testing.T) {
	s := acctcore.NewRuntimeBlockState(time.Now)
	a := &acctcore.Record{LoadLocation: time.LoadLocation, ID: 72, Platform: capability.PlatformOpenAI, Type: capability.ProviderTypeOAuth}
	s.BlockProviderScheduling(a, time.Now().Add(time.Minute), "first")
	fence := s.ManagedRecoveryFence(a.ID)
	s.BlockProviderScheduling(a, time.Now().Add(2*time.Minute), "new")
	require.False(t, s.ClearProviderSchedulingBlockIfFence(a.ID, fence))
	require.True(t, s.Blocked(a.ID, func() string { return acctcore.RefreshCredentialIdentity(a) }))
	require.True(t, s.ClearProviderSchedulingBlockIfFence(a.ID, s.ManagedRecoveryFence(a.ID)))
	require.False(t, s.Blocked(a.ID, func() string { return acctcore.RefreshCredentialIdentity(a) }))
	require.False(t, s.ClearProviderSchedulingBlockIfFence(a.ID, fence))
}
