package app

import "testing"

// TestProductEnvCompatibility 检查调试变量的旧名称兼容，以及新变量关闭时的优先级。
func TestProductEnvCompatibility(t *testing.T) {
	t.Setenv("SUB2API_DEBUG_MODEL_ROUTING", "true")
	t.Setenv("TOKENROUTER_DEBUG_MODEL_ROUTING", "")
	if !selectionOptions(nil).DebugRouting {
		t.Fatal("legacy routing flag was ignored")
	}
	for _, value := range []string{"0", "false"} {
		t.Setenv("TOKENROUTER_DEBUG_MODEL_ROUTING", value)
		if selectionOptions(nil).DebugRouting {
			t.Fatalf("new value %q must disable routing debug", value)
		}
	}
}
