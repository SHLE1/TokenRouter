package protocol

// TokenUsage 保存输入、输出和缓存 token 用量。
type TokenUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreation5mTokens    int // 5分钟缓存创建token（来自嵌套 cache_creation 对象）
	CacheCreation1hTokens    int // 1小时缓存创建token（来自嵌套 cache_creation 对象）
	ImageOutputTokens        int `json:"image_output_tokens,omitempty"`
	// Speed 记录 Claude 实际返回的处理速度，"fast" 会映射到内部 priority 计费。
	Speed string `json:"speed,omitempty"`
}

// HasObservedTokens 判断是否观测到正数 token 用量。
func (u *TokenUsage) HasObservedTokens() bool {
	if u == nil {
		return false
	}
	return u.InputTokens > 0 || u.OutputTokens > 0 ||
		u.CacheCreationInputTokens > 0 || u.CacheReadInputTokens > 0 ||
		u.CacheCreation5mTokens > 0 || u.CacheCreation1hTokens > 0 ||
		u.ImageOutputTokens > 0
}
