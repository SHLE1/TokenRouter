package httpapi

import (
	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	providercore "github.com/TokenFlux/TokenRouter/internal/provider"
	openaiws "github.com/TokenFlux/TokenRouter/internal/upstream/openai/ws"
)

func activeCodexFingerprintMode(provider *gatewayprovider.ExecutionProvider) providercore.CodexFingerprintMode {
	if provider == nil || gatewayprovider.ExecutionProtocolRecord(provider).GetCodexFingerprintMode() == providercore.CodexFingerprintOff {
		return providercore.CodexFingerprintOff
	}
	if _, ok := providercore.CodexFingerprintSeed(provider.Record.Extra); !ok {
		return providercore.CodexFingerprintOff
	}
	return gatewayprovider.ExecutionProtocolRecord(provider).GetCodexFingerprintMode()
}

func openAIWSPoolProviderView(provider *gatewayprovider.ExecutionProvider) *openaiws.WSPoolProvider {
	if provider == nil {
		return nil
	}
	return &openaiws.WSPoolProvider{ID: provider.Record.ID, Concurrency: provider.Record.Concurrency, Type: provider.Record.Type, FingerprintMode: string(activeCodexFingerprintMode(provider))}
}
