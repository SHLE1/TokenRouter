package provider

import "github.com/TokenFlux/TokenRouter/internal/gateway/provider/modelidentity"

// CompletionModels 提供用量结算时的候选型号。
type CompletionModels struct{}

func (CompletionModels) Candidates(model string, alternates ...string) []string {
	return modelidentity.UsageCandidates(model, alternates...)
}
