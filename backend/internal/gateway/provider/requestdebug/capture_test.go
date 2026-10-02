package requestdebug

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCaptureOnlyForRetainedDiagnosticOrEnabledLogging 检查诊断快照是否按调用方需要和日志开关生成。
func TestCaptureOnlyForRetainedDiagnosticOrEnabledLogging(t *testing.T) {
	request := httptest.NewRequest("POST", "https://example.test/v1/messages", nil)
	request.Header.Set("Authorization", "Bearer fixture-secret")
	trace := New("", "false")
	if line := trace.Capture(request, nil, nil, "apikey", false, false); line != "" {
		t.Fatalf("关闭诊断时不应生成快照：%s", line)
	}
	line := trace.Capture(request, nil, nil, "oauth", true, true)
	if line == "" || !strings.Contains(line, "[redacted]") || strings.Contains(line, "fixture-secret") {
		t.Fatalf("保留的诊断必须包含脱敏后的身份：%s", line)
	}
}
