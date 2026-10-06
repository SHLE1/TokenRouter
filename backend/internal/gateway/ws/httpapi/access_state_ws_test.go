package httpapi

import (
	"context"
	"testing"
	"time"

	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	providercore "github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
	"github.com/stretchr/testify/require"
)

func TestOpenAIWSStandaloneFailedStructured403AppliesProviderSideEffectsOnce(t *testing.T) {
	repo := &openAIStream403ProviderRepo{}
	svc := newWSFixture(wsFixtureInputs{health: newUpstreamHealthForTest(repo, nil, nil, providercore.HealthOptions{}, nil)})
	provider := &gatewayprovider.ExecutionProvider{Record: providercore.Record{LoadLocation: time.LoadLocation, ID: 923, Platform: capability.PlatformOpenAI, Type: capability.ProviderTypeOAuth}}
	failed := []byte(`{"type":"response.failed","response":{"error":{"type":"permission_error","code":"invalid_api_key","status_code":403,"message":"credential rejected"}}}`)

	require.True(t, svc.handleOpenAIWSFailureProviderSideEffects(context.Background(), provider, "gpt-5", nil, failed))
	require.Equal(t, 1, repo.setErrorCalls)
	require.True(t, wsFixtureProviderBlocked(svc, provider))
}

func TestOpenAIWSPairedStructured403SideEffectsCanBeDeduplicated(t *testing.T) {
	repo := &openAIStream403ProviderRepo{}
	svc := newWSFixture(wsFixtureInputs{health: newUpstreamHealthForTest(repo, nil, nil, providercore.HealthOptions{}, nil)})
	provider := &gatewayprovider.ExecutionProvider{Record: providercore.Record{LoadLocation: time.LoadLocation, ID: 924, Platform: capability.PlatformOpenAI, Type: capability.ProviderTypeOAuth}}
	errorEvent := []byte(`{"type":"error","error":{"code":"workspace_suspended","status_code":403,"message":"workspace is suspended"}}`)
	failedEvent := []byte(`{"type":"response.failed","response":{"error":{"code":"workspace_suspended","status_code":403,"message":"workspace is suspended"}}}`)

	applied := svc.handleOpenAIWSFailureProviderSideEffects(context.Background(), provider, "gpt-5", nil, errorEvent)
	if !applied {
		applied = svc.handleOpenAIWSFailureProviderSideEffects(context.Background(), provider, "gpt-5", nil, failedEvent)
	}

	require.True(t, applied)
	require.Equal(t, 1, repo.setErrorCalls)
}
