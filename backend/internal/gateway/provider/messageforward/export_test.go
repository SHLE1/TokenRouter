package messageforward

import (
	"context"
	"net/http"

	"github.com/TokenFlux/TokenRouter/internal/gateway/forward"
	gatewayadapter "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
	"github.com/TokenFlux/TokenRouter/internal/upstream/anthropic"
)

// ErrorForTest 等入口供外部测试调用 HTTP Adapter，随 go test 构建。
// 测试仍调用本包私有实现，生产 API 不增加白盒测试入口。
func ErrorForTest(runtime *Runtime, ctx context.Context, output HTTPBoundary, target *gatewayadapter.ExecutionProvider, response *http.Response, retry bool, models ...string) (*forward.Result, error) {
	return runtime.handleError(ctx, output, &AttemptState{}, target, response, retry, models...)
}

func ResponseOptionsForTest(runtime *Runtime, ctx context.Context, output HTTPBoundary, target *gatewayadapter.ExecutionProvider, model string, passthrough bool) anthropic.ResponseOptions {
	return runtime.responseOptions(ctx, output, &AttemptState{}, target, model, passthrough)
}
