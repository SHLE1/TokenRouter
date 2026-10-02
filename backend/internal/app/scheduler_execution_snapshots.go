package app

import (
	"github.com/TokenFlux/TokenRouter/internal/gateway/provider/selection"
	"github.com/TokenFlux/TokenRouter/internal/scheduler"
	schedulerredis "github.com/TokenFlux/TokenRouter/internal/scheduler/rediscache"
)

// provideSelectionSnapshots 返回快照读取适配器，来源缺失时返回 nil 接口。
func provideSelectionSnapshots(source *scheduler.SnapshotService) selection.Snapshots {
	if source == nil {
		return nil
	}
	return schedulerredis.NewSnapshotReader(source)
}
