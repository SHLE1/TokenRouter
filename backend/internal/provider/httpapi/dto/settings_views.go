package dto

// OverloadCooldownSettings 是过载冷却设置的 HTTP 数据结构。
type OverloadCooldownSettings struct {
	Enabled         bool `json:"enabled"`
	CooldownMinutes int  `json:"cooldown_minutes"`
}

// OpenAI403CooldownSettings 是 OpenAI 403 冷却设置的 HTTP 数据结构。
type OpenAI403CooldownSettings struct {
	Enabled                 bool `json:"enabled"`
	CooldownMinutes         int  `json:"cooldown_minutes"`
	ErrorOnThresholdEnabled bool `json:"error_on_threshold_enabled"`
	ThresholdCount          int  `json:"threshold_count"`
	ThresholdWindowMinutes  int  `json:"threshold_window_minutes"`
}

// RateLimit429CooldownSettings 是 429 冷却设置的 HTTP 数据结构。
type RateLimit429CooldownSettings struct {
	Enabled         bool `json:"enabled"`
	CooldownSeconds int  `json:"cooldown_seconds"`
}

// OpenAIImagesOAuthUnavailableCooldownSettings 是 OAuth 图片不可用时的冷却设置。
type OpenAIImagesOAuthUnavailableCooldownSettings struct {
	CooldownMinutes int `json:"cooldown_minutes"`
}

// OpenAIOAuthImportProviderDefaults 是 OAuth 导入模板中的提供商默认字段。
type OpenAIOAuthImportProviderDefaults struct {
	Notes              *string  `json:"notes,omitempty"`
	Concurrency        *int     `json:"concurrency,omitempty"`
	Priority           *int     `json:"priority,omitempty"`
	RateMultiplier     *float64 `json:"rate_multiplier,omitempty"`
	ExpiresAt          *int64   `json:"expires_at,omitempty"`
	AutoPauseOnExpired *bool    `json:"auto_pause_on_expired,omitempty"`
}

// OpenAIOAuthImportDefaults 是 OpenAI OAuth 导入默认设置。
type OpenAIOAuthImportDefaults struct {
	Provider    OpenAIOAuthImportProviderDefaults `json:"provider,omitempty"`
	Credentials map[string]any                    `json:"credentials,omitempty"`
	Extra       map[string]any                    `json:"extra,omitempty"`
}

// StreamTimeoutSettings 是流式请求超时设置的 HTTP 数据结构。
type StreamTimeoutSettings struct {
	Enabled                bool   `json:"enabled"`
	Action                 string `json:"action"`
	TempUnschedMinutes     int    `json:"temp_unsched_minutes"`
	ThresholdCount         int    `json:"threshold_count"`
	ThresholdWindowMinutes int    `json:"threshold_window_minutes"`
}
