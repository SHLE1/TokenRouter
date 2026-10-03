package wsentry

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TokenFlux/TokenRouter/internal/apikey"
	keyhttp "github.com/TokenFlux/TokenRouter/internal/apikey/httpapi"
	"github.com/TokenFlux/TokenRouter/internal/routing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type entryKeyReader struct{ calls int }

func (r *entryKeyReader) GetByKey(context.Context, string) (*apikey.APIKey, error) {
	r.calls++
	return nil, nil
}

func (r *entryKeyReader) Reauthenticate(context.Context, *apikey.APIKey, apikey.AuthenticationInput) (*apikey.APIKey, error) {
	r.calls++
	return nil, apikey.ErrAPIKeyNotFound
}

// AcquireRequest 为访问快照测试提供请求准入接口。
func (r *entryKeyReader) AcquireRequest(ctx context.Context, _ *apikey.APIKey) (context.Context, func(), time.Duration, error) {
	return ctx, func() {}, 0, nil
}

// TestEntryAccessKeepsProjectionIndependentAndLazy 验证策略按需刷新，模型和 effort 映射使用独立数据，认证快照保持原样。
func TestEntryAccessKeepsProjectionIndependentAndLazy(t *testing.T) {
	reader := &entryKeyReader{}
	key := &apikey.APIKey{ID: 9, UserID: 7, Key: "fixture-key", ModelMapping: map[string]string{"alias": "original"}, Group: &routing.Group{ReasoningEffortMappings: []routing.ReasoningEffortMapping{{}}}}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(string(keyhttp.ContextKeyAPIKey), key)
	view, ok := (openAIWSHTTPBackend{bindings: Bindings{Keys: reader}}).Access(c)
	require.True(t, ok)
	require.True(t, view.RefreshFastPolicy)
	require.Zero(t, reader.calls)
	view.ModelMapping["alias"] = "changed"
	require.Equal(t, "original", key.ModelMapping["alias"])
	require.NotSame(t, &key.Group.ReasoningEffortMappings[0], &view.Group.ReasoningEffortMappings[0])
}
