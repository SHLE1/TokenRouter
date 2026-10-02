package searchtools

// StandaloneRequest 保存独立搜索入口的请求字段，供平台构造器使用。
type StandaloneRequest struct {
	Query                    string   `json:"query"`
	Input                    string   `json:"input"`
	MaxResults               *int     `json:"max_results"`
	AllowedXHandles          []string `json:"allowed_x_handles"`
	ExcludedXHandles         []string `json:"excluded_x_handles"`
	FromDate                 string   `json:"from_date"`
	ToDate                   string   `json:"to_date"`
	EnableImageUnderstanding *bool    `json:"enable_image_understanding"`
	EnableVideoUnderstanding *bool    `json:"enable_video_understanding"`
}
