package httpapi

import (
	"net/http"
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/egress"
	egressadapter "github.com/TokenFlux/TokenRouter/internal/egress/provider"
)

// WriteAnthropicPassthroughHeaders 使用传入的过滤器，缺省时透传 Content-Type 和 x-request-id。
func WriteAnthropicPassthroughHeaders(dst, src http.Header, filter *egress.CompiledHeaderFilter) {
	if dst == nil || src == nil {
		return
	}
	if filter != nil {
		egressadapter.WriteFilteredHeaders(dst, src, filter)
		return
	}
	if value := strings.TrimSpace(src.Get("Content-Type")); value != "" {
		dst.Set("Content-Type", value)
	}
	if value := strings.TrimSpace(src.Get("x-request-id")); value != "" {
		dst.Set("x-request-id", value)
	}
}
