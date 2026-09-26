package httpapi

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// 管理接口明确拒绝已经移除的字段，避免旧表单静默丢失配置。
func bindGroupPlatformJSON(t *testing.T, target any, body string) error {
	t.Helper()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return bindManagementJSON(c, target)
}

func TestGroupManagementRejectsRetiredFields(t *testing.T) {
	for _, body := range []string{`{"name":"g","platform":"openai"}`, `{"name":"g","is_default":false}`} {
		require.Error(t, bindGroupPlatformJSON(t, &CreateGroupRequest{}, body))
		require.Error(t, bindGroupPlatformJSON(t, &UpdateGroupRequest{}, body))
	}
	var req CreateGroupRequest
	require.NoError(t, bindGroupPlatformJSON(t, &req, `{"name":"mixed","allowed_protocols":["anthropic_messages","openai_responses"],"protocol_fallbacks":{"anthropic_messages":[]}}`))
	require.Equal(t, "mixed", req.Name)
	require.NotNil(t, req.ProtocolFallbacks["anthropic_messages"])
}
