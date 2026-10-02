package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestModelMappingDefaultsAreLazyAndIsolated 检查有效映射直接返回，缺省映射按需读取并复制平台目录。
func TestModelMappingDefaultsAreLazyAndIsolated(t *testing.T) {
	calls := 0
	defaults := map[string]string{"alias": "default-model"}
	options := ModelMappingDefaults{Antigravity: func() map[string]string { calls++; return defaults }}
	r := &Record{Platform: PlatformAntigravity, Credentials: map[string]any{"model_mapping": map[string]any{"explicit": "target"}}}
	value := ResolveModelMapping(r, options)
	require.Zero(t, calls)
	require.Equal(t, "target", value["explicit"])
	r.Credentials = nil
	value = ResolveModelMapping(r, options)
	require.Equal(t, 1, calls)
	value["alias"] = "changed"
	require.Equal(t, "default-model", defaults["alias"])
	require.Equal(t, "default-model", ResolveModelMapping(r, options)["alias"])
}
