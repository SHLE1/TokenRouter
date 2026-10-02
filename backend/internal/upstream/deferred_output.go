package upstream

import "net/http"

// NewDeferredOutputContext 在首次访问输出接口时取得响应 Header。
func NewDeferredOutputContext(sink OutputSink) *OutputContext {
	return &OutputContext{Writer: &deferredOutputWriter{sink: sink}}
}

type deferredOutputWriter struct {
	sink   OutputSink
	output OutputWriter
}

func (w *deferredOutputWriter) writer() OutputWriter {
	if w.output == nil {
		w.output = NewOutputContext(w.sink).Writer
	}
	return w.output
}
func (w *deferredOutputWriter) Header() http.Header            { return w.writer().Header() }
func (w *deferredOutputWriter) Write(data []byte) (int, error) { return w.writer().Write(data) }
func (w *deferredOutputWriter) WriteHeader(status int)         { w.writer().WriteHeader(status) }
func (w *deferredOutputWriter) WriteHeaderNow()                { w.writer().WriteHeaderNow() }
func (w *deferredOutputWriter) Flush()                         { w.writer().Flush() }
func (w *deferredOutputWriter) Written() bool {
	if w.output != nil {
		return w.output.Written()
	}
	if state, ok := w.sink.(interface{ OutputState() OutputHead }); ok {
		return state.OutputState().Committed
	}
	return false
}
