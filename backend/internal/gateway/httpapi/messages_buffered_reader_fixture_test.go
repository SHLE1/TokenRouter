package httpapi

// messagesBufferedReadErrorFixture 返回读取失败信号并记录关闭结果。
type messagesBufferedReadErrorFixture struct{ err error }

func (r *messagesBufferedReadErrorFixture) Read([]byte) (int, error) { return 0, r.err }
func (r *messagesBufferedReadErrorFixture) Close() error             { return nil }
