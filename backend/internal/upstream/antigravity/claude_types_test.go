package antigravity

import "testing"

// TestReasoningRequestRules 覆盖需要特殊参数处理的完整型号。
func TestReasoningRequestRules(t *testing.T) {
	for _, id := range []string{"gemini-2.5-flash-thinking", "gemini-3-pro-high", "gemini-3.1-pro-high", "gemini-3.6-flash-high", "gemini-3.6-flash-low", "gemini-3.6-flash-medium", "gemini-3.6-flash-tiered"} {
		if !IsGeminiReasoningModel(id) {
			t.Errorf("missing request rule for %s", id)
		}
	}
	if IsGeminiReasoningModel("unknown-model") {
		t.Fatal("unknown model acquired reasoning rules")
	}
}
