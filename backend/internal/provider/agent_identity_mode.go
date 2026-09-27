// 本文件维护 provider 的所属能力；兼容入口复用唯一实现。
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
