package httpapi

import (
	"context"
	"time"

	gatewayprovider "github.com/TokenFlux/TokenRouter/internal/gateway/provider"
)

type transientCooldownProviderRepo struct {
	gatewayprovider.ExecutionProviderStore
}

func (transientCooldownProviderRepo) SetOverloaded(context.Context, int64, time.Time) error {
	return nil
}
