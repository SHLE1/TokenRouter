package httpapi

import (
	"github.com/TokenFlux/TokenRouter/internal/upstream/openai"
	"github.com/gin-gonic/gin"
)

const openAIResponsesRejectedFieldRetryBudgetContextKey = "openai_responses_rejected_field_retry_budget"

// openAIResponsesRejectedFieldRetryStateForRequest returns a fresh loop guard
// for one provider attempt backed by the inbound request's shared retry budget.
// A later provider may apply the same compatibility transform, while all provider
// attempts together remain bounded.
func openAIResponsesRejectedFieldRetryStateForRequest(c *gin.Context, initialBody []byte) *openai.ResponsesRejectedFieldRetryState {
	var budget *openai.ResponsesRejectedFieldRetryBudget
	if c != nil {
		if existing, ok := c.Get(openAIResponsesRejectedFieldRetryBudgetContextKey); ok {
			budget, _ = existing.(*openai.ResponsesRejectedFieldRetryBudget)
		}
	}
	if budget == nil {
		budget = &openai.ResponsesRejectedFieldRetryBudget{}
		if c != nil {
			c.Set(openAIResponsesRejectedFieldRetryBudgetContextKey, budget)
		}
	}
	return openai.NewOpenAIResponsesRejectedFieldRetryStateWithBudget(initialBody, budget)
}
