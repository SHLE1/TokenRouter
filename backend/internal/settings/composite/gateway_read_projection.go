package composite

import "github.com/TokenFlux/TokenRouter/internal/gateway"

// ApplyGatewayAdminReadSettings 将网关读取结果写入综合快照。
func (s *Snapshot) ApplyGatewayAdminReadSettings(value *gateway.AdminReadSettings) {
	s.AntigravityUserAgentVersion = value.AntigravityUserAgentVersion
	s.BackendModeEnabled = value.BackendModeEnabled
	s.ClaudeOAuthSystemPrompt = value.ClaudeOAuthSystemPrompt
	s.ClaudeOAuthSystemPromptBlocks = value.ClaudeOAuthSystemPromptBlocks
	s.EnableAnthropicCacheTTL1hInjection = value.EnableAnthropicCacheTTL1hInjection
	s.EnableCCHSigning = value.EnableCCHSigning
	s.EnableClaudeOAuthSystemPromptInjection = value.EnableClaudeOAuthSystemPromptInjection
	s.EnableClientDatelineNormalization = value.EnableClientDatelineNormalization
	s.EnableFingerprintUnification = value.EnableFingerprintUnification
	s.EnableIdentityPatch = value.EnableIdentityPatch
	s.EnableMetadataPassthrough = value.EnableMetadataPassthrough
	s.GrokDefaultBaseURLMode = value.GrokDefaultBaseURLMode
	s.GrokDefaultTextModel = value.GrokDefaultTextModel
	s.IdentityPatchPrompt = value.IdentityPatchPrompt
	s.MaxClaudeCodeVersion = value.MaxClaudeCodeVersion
	s.MinClaudeCodeVersion = value.MinClaudeCodeVersion
	s.OpenAIAllowClaudeCodeCodexPlugin = value.OpenAIAllowClaudeCodeCodexPlugin
	s.OpenAICodexUserAgent = value.OpenAICodexUserAgent
	s.OpenAITTFTMode = value.OpenAITTFTMode
	s.RewriteMessageCacheControl = value.RewriteMessageCacheControl
	s.UserPromptReplacementConfig = value.UserPromptReplacementConfig
}
