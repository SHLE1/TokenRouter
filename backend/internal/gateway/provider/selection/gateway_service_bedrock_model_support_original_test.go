package selection

import (
	"testing"
	"time"

	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	providercore "github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/routing/capability"
)

func TestGatewayServiceIsModelSupportedByProvider_BedrockDefaultMappingRestrictsModels(t *testing.T) {
	svc := newGenericSelectionForTest(GenericDependencies{Reads: Reads{}, Shared: Shared{}}, nil)

	provider := &gatewayprovider.ExecutionProvider{
		Record: providercore.Record{
			LoadLocation: time.LoadLocation, Platform: capability.PlatformAnthropic,
			Type: capability.ProviderTypeBedrock,
			Credentials: map[string]any{
				"aws_region": "us-east-1",
			},
		},
	}

	if !svc.isModelSupportedByProvider(provider, "claude-sonnet-4-5") {
		t.Fatalf("expected default Bedrock alias to be supported")
	}

	if svc.isModelSupportedByProvider(provider, "claude-3-5-sonnet-20241022") {
		t.Fatalf("expected unsupported alias to be rejected for Bedrock provider")
	}
}

func TestGatewayServiceIsModelSupportedByProvider_BedrockCustomMappingStillActsAsAllowlist(t *testing.T) {
	svc := newGenericSelectionForTest(GenericDependencies{Reads: Reads{}, Shared: Shared{}}, nil)

	provider := &gatewayprovider.ExecutionProvider{
		Record: providercore.Record{
			LoadLocation: time.LoadLocation, Platform: capability.PlatformAnthropic,
			Type: capability.ProviderTypeBedrock,
			Credentials: map[string]any{
				"aws_region": "eu-west-1",
				"model_mapping": map[string]any{
					"claude-sonnet-*": "claude-sonnet-4-6",
				},
			},
		},
	}

	if !svc.isModelSupportedByProvider(provider, "claude-sonnet-4-6") {
		t.Fatalf("expected matched custom mapping to be supported")
	}

	if !svc.isModelSupportedByProvider(provider, "claude-opus-4-6") {
		t.Fatalf("expected default Bedrock alias fallback to remain supported")
	}

	if svc.isModelSupportedByProvider(provider, "claude-3-5-sonnet-20241022") {
		t.Fatalf("expected unsupported model to still be rejected")
	}
}
