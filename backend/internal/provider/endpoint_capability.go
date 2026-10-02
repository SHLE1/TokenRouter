package provider

type OpenAIEndpointCapability string

const (
	OpenAIEndpointCapabilityTextGeneration OpenAIEndpointCapability = "text_generation"
	OpenAIEndpointCapabilityEmbeddings     OpenAIEndpointCapability = "embeddings"
	OpenAIEndpointCapabilityAlphaSearch    OpenAIEndpointCapability = "alpha_search"
	// OpenAIEndpointCapabilityLive 表示仅 ChatGPT OAuth 提供商支持的 Frameless Live 能力。
	OpenAIEndpointCapabilityLive OpenAIEndpointCapability = "live"
	// OpenAIEndpointCapabilityGrokMediaGeneration 用于排除已禁用或计费资格
	// 探测遭拒的 Grok 提供商；视频状态查询不要求该能力，以便继续查询已提交的任务。
	OpenAIEndpointCapabilityGrokMediaGeneration OpenAIEndpointCapability = "grok_media_generation"
	// OpenAIEndpointCapabilityResponses 表示上游确实提供 /v1/responses 端点。
	// 根据已启用的协议集合和路由规则判断。
	OpenAIEndpointCapabilityResponses OpenAIEndpointCapability = "responses"
	// OpenAIEndpointCapabilityRemoteCompactionV2 表示提供商可承接原生 remote_compaction_v2。
	// 它仍要求普通 Responses 能力，额外受提供商级 V2 开关控制。
	OpenAIEndpointCapabilityRemoteCompactionV2 OpenAIEndpointCapability = "remote_compaction_v2"
)
