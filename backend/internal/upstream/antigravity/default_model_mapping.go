package antigravity

// AntigravityGemini31ProAgentModel 是 Gemini 3.1 Pro High 的上游路由键。
const AntigravityGemini31ProAgentModel = "gemini-pro-agent"

// DefaultAntigravityModelMapping 将公开请求型号转换为平台内部路由键。
var DefaultAntigravityModelMapping = map[string]string{
	"gemini-3.1-pro-high": AntigravityGemini31ProAgentModel,
}
