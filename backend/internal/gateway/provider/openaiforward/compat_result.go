package openaiforward

import "github.com/TokenFlux/TokenRouter/internal/upstream/openai"

// FromCompatResult 将兼容响应结果转换为转发结果并附加计费模型。
func FromCompatResult(r *openai.CompatResponseResult, billing string) *Result {
	if r == nil {
		return nil
	}
	return &Result{RequestID: r.RequestID, ResponseID: r.ResponseID, Headers: r.UpstreamHeaders, Usage: r.Usage, Model: r.Model, BillingModel: billing, UpstreamModel: r.UpstreamModel, UpstreamResponseServiceTier: r.ServiceTier, ServiceTier: r.ResolvedTier, ReasoningEffort: r.ReasoningEffort, Stream: r.Stream, Duration: r.Duration, FirstTokenMs: r.FirstTokenMs, ClientDisconnect: r.ClientDisconnect, SearchCount: r.SearchCount}
}
