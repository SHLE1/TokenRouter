package selection

import "github.com/TokenFlux/TokenRouter/internal/scheduler"

// loads 为调度测试构造候选负载，选择器执行评分和槽位获取。
func (g *projectionScope) loads(values []providerWithLoad) []scheduler.FlowLoad {
	if values == nil {
		return nil
	}
	out := make([]scheduler.FlowLoad, len(values))
	for i, a := range values {
		out[i] = scheduler.FlowLoad{Provider: g.provider(a.provider), LoadInfo: a.loadInfo}
	}
	return out
}
