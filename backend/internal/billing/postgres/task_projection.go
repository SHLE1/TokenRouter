package postgres

import (
	"context"
	"database/sql"

	"github.com/TokenFlux/TokenRouter/internal/billing"
)

// @project-doc docs/domains/routing_and_billing.md#usage_settlement
type TaskProjection interface {
	SaveReservation(context.Context, float64, []billing.BillingAllocation, float64, float64) error
	SetAllowanceReserved(context.Context, bool) error
}
type (
	TaskProjectionFactory   func(*sql.Tx, billing.TaskReference) TaskProjection
	TaskProjectionFactories map[billing.TaskScope]TaskProjectionFactory
)
