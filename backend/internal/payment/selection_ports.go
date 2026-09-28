package payment

import (
	"context"
	"time"
)

type InstanceSource interface {
	EnabledInstances(context.Context, string) ([]*ProviderInstance, error)
	Instance(context.Context, int64) (*ProviderInstance, error)
	DailyUsage(context.Context, []string, time.Time) (map[string]float64, error)
	PaidDailyAmount(context.Context, string, time.Time) (float64, error)
}
type (
	SelectionObserver func(string, string, ...any)
	SelectionRuntime  struct {
		Now     func() time.Time
		Observe SelectionObserver
	}
)
