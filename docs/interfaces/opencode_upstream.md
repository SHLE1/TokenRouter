# OpenCode 上游

本文记录 OpenCode 提供商的模式、模型协议、转发和 Go 用量查询。分组的协议准入与转换规则见[统一协议能力](protocol_capabilities.md)，查询生命周期见[API Key 上游用量查询](upstream_usage.md)。

<a id="opencode_routing"></a>
## 模式与模型协议

提供商使用 `platform=opencode`、`type=apikey`。`credentials.api_key` 保存凭据；`provider_mode` 接受 `payg` 和 `go`，缺省为 `payg`。按量模式使用 Console / Zen，Go 使用订阅入口。

提供商可启用 Messages、Responses、Chat Completions、Gemini GenerateContent 和 SystemOne。实际候选还要检查最终上游模型的原生协议。例如 models.dev 内嵌目录中的 `minimax-m3` 在 payg 使用 Chat，在 Go 使用 Messages。模型名按完整 ID 精确匹配，提供商映射执行后才选协议。

`credentials.opencode_model_protocols` 保存两种模式各自的模型覆盖：

```json
{
  "payg": { "custom-model": "openai_responses" },
  "go": { "custom-model": "anthropic_messages" }
}
```

保存的覆盖优先于内嵌 models.dev 的协议目录。模型同步读取上游模型列表，并使用共享 models.dev 当前发布的协议目录补充建议；已保存的覆盖继续优先。同步结果进入表单，管理员保存后进入提供商配置。未识别的模型需要手动指定协议，无法确定协议的候选不参与调度。

目录协议索引独立于价格和属性的原厂筛选。它使用服务商或模型声明的 SDK 和 decision 类型，不根据名称猜测协议，也不把 OpenCode 的渠道价格加入默认价卡。

提供商的 `upstream_protocols` 继续限制可用协议，空集合关闭新调用。分组准入与已有单步转换规则在选路时生效；SystemOne 保持独立的同步 JSON 入口。模型目录、调度快照、换号复核和提供商测试使用同一份最终模型协议规则。

## 地址与认证

默认服务根地址为 `https://opencode.ai/inference`，Go 为 `https://opencode.ai/inference/go`。`credentials.base_url` 可以覆盖服务根地址；模型同步请求其 `/v1/models`。

| 协议 | 根地址后的默认路径 | `api_base_urls` 地址槽 |
| --- | --- | --- |
| Responses | `/openai/v1/responses` | `responses` |
| Chat | `/openai/v1/chat/completions` | `chat_completions` |
| Messages | `/anthropic/v1/messages` | `anthropic` |
| Gemini | `/google/v1beta/models/{model}:generateContent` | `gemini` |
| SystemOne | `/systemone/v1/systemone` | `systemone` |

地址槽保存完整协议根地址，包括 `/v1` 或 `/v1beta`；Gemini 流式请求使用 `:streamGenerateContent`。覆盖地址同样经过出站 URL 校验，并复用提供商代理和对应传输配置。

各协议使用 `Authorization: Bearer <api_key>`。发送前移除其他协议的 API Key 认证头。客户端 User-Agent 优先于提供商配置，缺失时使用配置值或 `TokenRouter`。会话头依次读取 `x-opencode-session`、`session-id`、`session_id`、`conversation_id`、`x-session-affinity`、`x-session-id`，发送为 `X-OpenCode-Session`；缺少稳定会话信息时不生成随机会话 ID。

OpenCode 的平台身份贯穿使用记录、错误策略和运行监控。兼容请求复用现有协议编解码、工具调用、流式处理和用量采集。SystemOne 不接受 `stream=true`。是否可调用某个上游模型仍受其账户权限和服务限制约束。

## 管理与用量

创建和编辑页面支持两种模式的模型协议覆盖、各协议地址、模型同步和测试。同步返回保留 `models` 数组，并为 OpenCode 增加 `model_protocols` 对象。预览请求接受 `provider_mode`；已保存提供商的同步请求可携带临时 `provider_mode` 和 `base_url`，查询使用服务器保存的 API Key，配置仍由管理员另行保存。

Go 使用固定适配器 `opencode_go`，请求 `GET https://opencode.ai/zen/go/v1/usage`，以提供商 API Key 做 Bearer 认证。查询根地址可通过现有用量配置覆盖。结果包含滚动五小时、周和月窗口的使用百分比、状态及 UTC 重置时间。超过 100% 的已用百分比保留，剩余额度最低为零。

三个窗口均需包含有效的 `percent`、`status` 和 `resetsAt`。异常响应、认证错误、未订阅、限流和超时按查询失败展示。单个及批量查询沿用身份复核、独立并发和五分钟浏览器缓存；查询失败不改变转发资格。

payg 不提供上游用量查询，列表仍可展示本地统计。此接入没有钱包余额查询或后台额度监控。用量查询不会根据窗口百分比自动停调，也不会修改 OpenCode 控制台的超额使用设置。

参考：[官方推理 API](https://opencode.ai/v2/docs/console/api/inference/)、[Go](https://opencode.ai/v2/docs/console/go/)、[官方 Go 查询实现](https://github.com/anomalyco/opencode/blob/dev/packages/core/src/usage-limits.ts)。接入协议核验日期：2026-10-26。
