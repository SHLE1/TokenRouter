package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// TestProviderImportsRejectRetiredGroupInputs 检查 OAuth 和导入入口在创建提供商或交换凭据前拒绝废弃字段。
func TestProviderImportsRejectRetiredGroupInputs(t *testing.T) {
	router := gin.New()
	router.POST("/batch", (&ManagementHandler{}).BatchCreate)
	router.POST("/archive", (&ArchiveHandler{}).ImportData)
	router.POST("/codex", (&CodexImportHandler{}).ImportCodexSession)
	router.POST("/pat", (&OpenAIOAuthHandler{}).CreateProviderFromCodexPAT)
	router.POST("/oauth", (&OpenAIOAuthHandler{}).CreateProviderFromOAuth)
	router.POST("/shadow/:id", (&OpenAIOAuthHandler{}).CreateShadow)
	router.POST("/grok/oauth", (&GrokOAuthHandler{}).CreateProviderFromOAuth)
	router.POST("/grok/sso", (&GrokOAuthHandler{}).CreateProvidersFromSSO)
	for _, field := range []string{"confirm_mixed_channel_risk", "skip_default_group_bind"} {
		for _, path := range []string{"/archive", "/codex", "/pat", "/oauth", "/shadow/1", "/grok/oauth", "/grok/sso", "/batch"} {
			t.Run(path+field, func(t *testing.T) {
				body := `{"` + field + `":null}`
				if path == "/batch" {
					body = `{"providers":[{"name":"a","platform":"openai","type":"apikey","credentials":{},"` + field + `":false}]}`
				}
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, req)
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				require.Contains(t, recorder.Body.String(), field)
			})
		}
	}
}
