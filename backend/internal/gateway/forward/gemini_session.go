package forward

// GeminiSession 固化执行入口提供的会话归属，不负责选号或创建远端会话。
type GeminiSession struct {
	GroupID     int64
	SessionHash string
}

// GeminiSessionOption 可省略或传 nil，多个选项按顺序覆盖。
type GeminiSessionOption func(*GeminiSession)

func WithGeminiSession(groupID int64, sessionHash string) GeminiSessionOption {
	return func(options *GeminiSession) { options.GroupID = groupID; options.SessionHash = sessionHash }
}
