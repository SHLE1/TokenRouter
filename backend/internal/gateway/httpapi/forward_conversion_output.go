package httpapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/TokenFlux/TokenRouter/internal/egress"
	egressprovider "github.com/TokenFlux/TokenRouter/internal/egress/provider"
	forwardcore "github.com/TokenFlux/TokenRouter/internal/gateway/forward"
	protocolopenai "github.com/TokenFlux/TokenRouter/internal/protocol/openai"
	"github.com/gin-gonic/gin"
)

// ForwardConversionOutput 保存当前 HTTP 输出和协议转换回调。
type ForwardConversionOutput struct {
	Context    *gin.Context
	Filter     *egress.CompiledHeaderFilter
	Responses  bool
	Reverse    func([]byte) []byte
	Commit     func()
	Diagnostic func(string, string, error, string, string)
}

func (o ForwardConversionOutput) CopyHeaders(headers map[string][]string) {
	if o.Filter != nil {
		egressprovider.WriteFilteredHeaders(o.Context.Writer.Header(), headers, o.Filter)
	}
}

func (o ForwardConversionOutput) BeginJSON() {
	o.Context.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func (o ForwardConversionOutput) BeginStream() {
	h := o.Context.Writer.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
	o.Context.Writer.WriteHeader(http.StatusOK)
}
func (o ForwardConversionOutput) ReverseTools(body []byte) []byte { return o.Reverse(body) }
func (o ForwardConversionOutput) JSONBytes(body []byte) {
	o.Context.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

func (o ForwardConversionOutput) ResponsesJSON(value *protocolopenai.ResponsesResponse) {
	o.Context.JSON(http.StatusOK, value)
}

func (o ForwardConversionOutput) ChatJSON(value *protocolopenai.ChatCompletionsResponse) {
	o.Context.JSON(http.StatusOK, value)
}

func (o ForwardConversionOutput) Event(kind string, body []byte) (int, error) {
	if kind != "" {
		return fmt.Fprintf(o.Context.Writer, "event: %s\ndata: %s\n\n", kind, body)
	}
	return fmt.Fprintf(o.Context.Writer, "data: %s\n\n", body)
}
func (o ForwardConversionOutput) Flush() { o.Context.Writer.Flush() }
func (o ForwardConversionOutput) Error(status int, kind, message string) {
	o.Commit()
	if o.Responses {
		o.Context.JSON(status, gin.H{"error": gin.H{"code": kind, "message": message}})
		return
	}
	o.Context.JSON(status, gin.H{"error": gin.H{"type": kind, "message": message}})
}

func (o ForwardConversionOutput) Observe(level, message string, err error, requestID, event string) {
	o.Diagnostic(level, message, err, requestID, event)
}

var _ forwardcore.Output = ForwardConversionOutput{}

// WriteForwardMessageGenericError 标记响应提交，并写出 Messages 通用错误。
func WriteForwardMessageGenericError(c *gin.Context, commit func()) {
	commit()
	c.JSON(http.StatusInternalServerError, gin.H{"type": "error", "error": gin.H{"type": "upstream_error", "message": "Upstream gateway error"}})
}

// WriteForwardCountSuccess 在透传时使用上游 Content-Type，其他计数响应使用 JSON。
func WriteForwardCountSuccess(c *gin.Context, status int, headers map[string][]string, body []byte, passthrough bool) {
	contentType := "application/json"
	if passthrough {
		if value := strings.TrimSpace(http.Header(headers).Get("Content-Type")); value != "" {
			contentType = value
		}
	}
	c.Data(status, contentType, body)
}

// WriteForwardCountError 按 Anthropic 格式写出计数错误，提交标记由调用方管理。
func WriteForwardCountError(c *gin.Context, status int, kind, message string) {
	c.JSON(status, gin.H{"type": "error", "error": gin.H{"type": kind, "message": message}})
}

// WriteForwardGeminiErrorBody 写出调用方传入的 Gemini 错误。
func WriteForwardGeminiErrorBody(c *gin.Context, status int, contentType string, body []byte, commit func()) {
	commit()
	c.Data(status, contentType, body)
}
