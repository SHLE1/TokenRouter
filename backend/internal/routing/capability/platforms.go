package capability

// AccountPlatforms 返回账号平台目录的独立副本，供跨平台分组构造候选池。
func AccountPlatforms() []string {
	return []string{PlatformAnthropic, PlatformOpenAI, PlatformGemini, PlatformAntigravity, PlatformGrok, PlatformQoder, PlatformKimi, PlatformZhipu, PlatformDeepseek}
}
