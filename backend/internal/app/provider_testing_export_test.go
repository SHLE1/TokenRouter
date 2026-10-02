//go:build integration

package app

// NewProviderTestsForTest 供集成测试调用提供商测试组件的应用构造函数。
var NewProviderTestsForTest = provideProviderTests

// 这些构造函数供集成测试组合平台探测和应用任务跟踪器。
var (
	NewAntigravityRetryForTest = provideAntigravityRetry
	NewAntigravityProbeForTest = provideAntigravityProbe
	NewGatewayActivityForTest  = provideGatewayRequestActivity
)

// 快照回放测试使用生产装配和 outbox 发布器。
var (
	NewSnapshotForTest       = provideSchedulerSnapshot
	NewProviderEventsForTest = newProviderEvents
)

// 健康状态集成测试使用应用构造函数。
var (
	NewProviderHealthRuntimeForTest = provideProviderHealthRuntime
	NewUpstreamHealthForTest        = provideUpstreamHealth
)

// 完成记录集成测试使用应用中的记录器和倍率缓存。
var (
	NewCompletionRecordersForTest = ProvideGatewayCompletionRecorders
	NewGatewayBillingRatesForTest = provideGatewayBillingRates
)

// NewExecutionProviderStoreForTest 为集成测试构造执行提供商存储。
var NewExecutionProviderStoreForTest = provideExecutionProviderStore

// NewProviderStoreForTest 为集成测试绑定提供商存储配置和事件写入函数。
var NewProviderStoreForTest = provideProviderStore

// NewModelAvailabilityForTest 为集成测试构造模型诊断组件。
var NewModelAvailabilityForTest = provideGatewayModelAvailability
