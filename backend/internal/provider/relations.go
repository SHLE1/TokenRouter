package provider

import (
	"time"

	"github.com/TokenFlux/TokenRouter/internal/routing/accessview"
)

// GroupMembership 保存提供商与分组的关联。
type GroupMembership struct {
	ProviderID int64
	GroupID    int64
	CreatedAt  time.Time
	Provider   *Record
	Group      *accessview.GroupConfig
}
