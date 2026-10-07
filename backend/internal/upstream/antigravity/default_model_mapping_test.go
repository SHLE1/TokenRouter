package antigravity

import "testing"

// TestDefaultMappingOnlyRewritesRouteKeys 自映射不再充当默认白名单。
func TestDefaultMappingOnlyRewritesRouteKeys(t *testing.T) {
	if len(DefaultAntigravityModelMapping) != 1 || DefaultAntigravityModelMapping["gemini-3.1-pro-high"] != AntigravityGemini31ProAgentModel {
		t.Fatal("unexpected execution mapping")
	}
}
