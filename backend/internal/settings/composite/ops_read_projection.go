package composite

import "github.com/TokenFlux/TokenRouter/internal/ops"

// ApplyOpsAdminReadSettings 将运维读取结果写入综合快照。
func (s *Snapshot) ApplyOpsAdminReadSettings(value *ops.AdminReadSettings) {
	s.OpenAIQuotaAutoPauseSettings = value.OpenAIQuotaAutoPauseSettings
	s.OpsMetricsIntervalSeconds = value.OpsMetricsIntervalSeconds
	s.OpsMonitoringEnabled = value.OpsMonitoringEnabled
	s.OpsRealtimeMonitoringEnabled = value.OpsRealtimeMonitoringEnabled
}
