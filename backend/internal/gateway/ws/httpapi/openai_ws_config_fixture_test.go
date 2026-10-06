package httpapi

// newOpenAIWSV2TestConfig 为网络夹具设置短退避，算法测试自行指定时长。
func newOpenAIWSV2TestConfig() *wsFixtureOptions {
	options := &wsFixtureOptions{}

	options.WS.StickyResponseIDTTLSeconds = 3600

	return options
}
