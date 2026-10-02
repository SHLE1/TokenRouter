package upstream

import (
	"encoding/json"
	"io"
	"net/http"
)

// OutputHead 是输出适配器需要的响应元数据，Header 在传递时复制。
type OutputHead struct {
	// Committed 记录输出适配器当前的 HTTP 提交状态。
	Committed bool
	Status    int
	Header    http.Header
}

// OutputEvent 是一个输出片段，写入和刷新错误同步返回执行方。
type OutputEvent struct {
	Data           []byte
	Flush          bool
	Semantic       bool
	CommitForRetry bool
	Terminal       bool
}

// OutputSink 由 HTTP 或其他调用适配器实现，拥有实际写入与刷新。
type OutputSink interface {
	Begin(OutputHead) error
	Emit(OutputEvent) error
}

// OutputContext 提供一次转换所需的字节输出接口。
type OutputContext struct{ Writer OutputWriter }

// OutputWriter 按编解码器的调用顺序，通过 sink 写出数据。
type OutputWriter interface {
	io.Writer
	Header() http.Header
	WriteHeader(int)
	WriteHeaderNow()
	Written() bool
	Flush()
}

// NewOutputContext 为一次流转换建立独立元数据，禁止跨请求复用。
func NewOutputContext(sink OutputSink) *OutputContext {
	header := make(http.Header)
	status := http.StatusOK
	committed := false
	if source, ok := sink.(interface{ InitialOutput() OutputHead }); ok {
		head := source.InitialOutput()
		for key, values := range head.Header {
			header[key] = append([]string(nil), values...)
		}
		committed = head.Committed
		if committed && head.Status != 0 {
			status = head.Status
		}
	}
	return &OutputContext{Writer: &sinkWriter{sink: sink, header: header, status: status, written: committed}}
}

type sinkWriter struct {
	sink    OutputSink
	header  http.Header
	status  int
	started bool
	written bool
	err     error
	next    *OutputEvent
}

func (w *sinkWriter) Header() http.Header { return w.header }
func (w *sinkWriter) Written() bool       { return w.written }
func (w *sinkWriter) WriteHeader(status int) {
	if !w.started {
		w.status = status
	}
}

func (w *sinkWriter) begin() error {
	if w.err != nil {
		return w.err
	}
	if !w.started {
		w.started = true
		w.err = w.sink.Begin(OutputHead{Status: w.status, Header: w.header.Clone()})
		if w.err == nil {
			w.written = true
		}
	}
	return w.err
}

func (w *sinkWriter) Write(p []byte) (int, error) {
	if err := w.begin(); err != nil {
		return 0, err
	}
	w.written = true
	event := OutputEvent{CommitForRetry: true}
	if w.next != nil {
		event = *w.next
		w.next = nil
	}
	event.Data = p
	w.err = w.sink.Emit(event)
	if w.err != nil {
		return 0, w.err
	}
	return len(p), nil
}

func (w *sinkWriter) Flush() {
	if w.begin() == nil {
		w.err = w.sink.Emit(OutputEvent{Flush: true})
	}
}

// Header 只修改待发元数据，实际响应仍由 OutputSink 写入。
func (c *OutputContext) Header(key, value string) { c.Writer.Header().Set(key, value) }

// WriteHeaderNow 通过 sink.Begin 提交响应头，刷新由 Flush 触发。
func (w *sinkWriter) WriteHeaderNow() { _ = w.begin() }

// Data 设置缺省内容类型并写出响应，无 body 状态提交响应头。
func (c *OutputContext) Data(status int, contentType string, body []byte) {
	if len(c.Writer.Header()["Content-Type"]) == 0 {
		c.Writer.Header()["Content-Type"] = []string{contentType}
	}
	c.Writer.WriteHeader(status)
	if status < 200 || status == http.StatusNoContent || status == http.StatusNotModified {
		c.Writer.WriteHeaderNow()
		return
	}
	_, _ = c.Writer.Write(body)
}

// NextEvent 为下一次同步字节写入补充协议事实，不改变提交和刷新时点。
func (c *OutputContext) NextEvent(semantic, terminal bool) {
	if writer, ok := c.Writer.(*deferredOutputWriter); ok {
		(&OutputContext{Writer: writer.writer()}).NextEvent(semantic, terminal)
		return
	}
	if writer, ok := c.Writer.(*sinkWriter); ok {
		writer.next = &OutputEvent{Semantic: semantic, Terminal: terminal, CommitForRetry: true}
	}
}

// Status 设置后续输出使用的状态码，响应在写入时提交。
func (c *OutputContext) Status(status int) { c.Writer.WriteHeader(status) }

// JSON 编码并通过 Data 输出响应，编码失败时提交状态码。
func (c *OutputContext) JSON(status int, value any) {
	body, err := json.Marshal(value)
	if err != nil {
		c.Writer.WriteHeader(status)
		c.Writer.WriteHeaderNow()
		return
	}
	c.Data(status, "application/json; charset=utf-8", body)
}

// Err 返回本次同步输出的错误，后续读取或取消由平台决定。
func (c *OutputContext) Err() error {
	if writer, ok := c.Writer.(*deferredOutputWriter); ok {
		if writer.output == nil {
			return nil
		}
		return (&OutputContext{Writer: writer.output}).Err()
	}
	if writer, ok := c.Writer.(*sinkWriter); ok {
		return writer.err
	}
	return nil
}
