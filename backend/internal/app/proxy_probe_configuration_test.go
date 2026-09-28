package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/config"
	"github.com/stretchr/testify/require"
)

// TestNewProxyExitInfoProberUsesConfiguredTargets 验证配置必须控制实际请求目标，装配不能忽略配置而使用内置地址。
func TestNewProxyExitInfoProberUsesConfiguredTargets(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"ip":"203.0.113.42"}`))
	}))
	defer server.Close()
	prober := provideEgressProbe(&config.Config{Security: config.SecurityConfig{ProxyProbe: config.ProxyProbeConfig{URLs: []config.ProbeURLConfig{{URL: server.URL, Parser: "ipify"}}}}})
	result, _, err := prober.ProbeProxy(context.Background(), "")
	require.NoError(t, err)
	require.Equal(t, "203.0.113.42", result.IP)
	require.Equal(t, 1, calls)
}
