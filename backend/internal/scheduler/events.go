package scheduler

const (
	SchedulerOutboxEventProviderChanged       = "provider_changed"
	SchedulerOutboxEventProviderGroupsChanged = "provider_groups_changed"
	SchedulerOutboxEventProviderBulkChanged   = "provider_bulk_changed"
	SchedulerOutboxEventProviderLastUsed      = "provider_last_used"
	SchedulerOutboxEventGroupChanged          = "group_changed"
	SchedulerOutboxEventFullRebuild           = "full_rebuild"
)

// GroupPayload 在分组为空时返回 untyped nil，该值参与持久化去重指纹计算。
func GroupPayload(groupIDs []int64) any {
	if len(groupIDs) == 0 {
		return nil
	}
	return map[string]any{"group_ids": groupIDs}
}
