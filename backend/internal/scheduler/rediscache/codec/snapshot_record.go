package codec

import (
	"fmt"
	"slices"

	"github.com/TokenFlux/TokenRouter/internal/provider"
	"github.com/TokenFlux/TokenRouter/internal/scheduler"
)

// recordSnapshot 包装提供商记录，调度器通过 SnapshotMetadata 读取重建元数据。
// 执行适配器通过 RecordValue 取得含凭据的完整记录。
type recordSnapshot struct{ value *provider.Record }

func (s recordSnapshot) SnapshotMetadata() scheduler.SnapshotMetadata {
	return scheduler.SnapshotMetadata{
		ID: s.value.ID, Name: s.value.Name, Platform: s.value.Platform,
		GroupIDs: slices.Clone(s.value.GroupIDs),
	}
}

func WrapRecord(value *provider.Record) scheduler.SnapshotProvider {
	if value == nil {
		return nil
	}
	return recordSnapshot{value: value}
}

// RecordValue 为提供商读取和缓存编码返回完整记录，记录含有执行凭据。
func RecordValue(value scheduler.SnapshotProvider) (*provider.Record, error) {
	if value == nil {
		return nil, nil
	}
	stored, ok := value.(recordSnapshot)
	if !ok {
		return nil, fmt.Errorf("unexpected legacy scheduler snapshot data %T", value)
	}
	return stored.value, nil
}
