//go:build integration

package app

// NewS16ProviderTests 仅供外部集成测试调用真实组合根，不扩大生产 API。
var NewS16ProviderTests = provideProviderTests

// 以下入口仅在集成测试中组合真实平台探测与应用关闭屏障。
var (
	NewS16AntigravityRetry = provideAntigravityRetry
	NewS16AntigravityProbe = provideAntigravityProbe
	NewS16GatewayActivity  = provideGatewayRequestActivity
)

// 快照回放测试复用生产装配与原 outbox 发布器。
var (
	NewS16Snapshot       = provideSchedulerSnapshot
	NewS16ProviderEvents = newProviderEvents
)

// 集成测试直接使用原生装配及单向兼容绑定。
var (
	NewS16ProviderHealthRuntime = provideProviderHealthRuntime
	S16UpstreamHealth           = provideUpstreamHealth
)

// 原生完成装配仅向隔离存储测试开放，不增加生产 API。
var (
	NewS16CompletionRecorders = ProvideGatewayCompletionRecorders
	NewS16GatewayBillingRates = provideGatewayBillingRates
)

// 执行提供商集成合同通过真实装配绑定相同存储，不扩大生产接口。
var NewS16ExecutionProviderStore = provideExecutionProviderStore

// 提供商存储合同复用生产配置投影与事件绑定。
var NewS16ProviderStore = provideProviderStore

// 模型诊断集成合同直接使用实际组合根，不构造旧执行服务。
var S16ModelAvailability = provideGatewayModelAvailability
