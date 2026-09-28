package provider

import (
	"strings"
)

const OpenAIAuthModeAgentIdentity = "agentIdentity"

func (a *Record) IsOpenAIAgentIdentity() bool {
	if a == nil || !a.IsOpenAIOAuth() {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(a.GetCredential(OpenAIAuthModeCredentialKey)), OpenAIAuthModeAgentIdentity)
}
