package payment

import (
	"context"
	"time"
)

type AuditLog struct {
	ID        int64     `json:"id,omitempty"`
	OrderID   string    `json:"order_id,omitempty"`
	Action    string    `json:"action,omitempty"`
	Detail    string    `json:"detail,omitempty"`
	Operator  string    `json:"operator,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

func (s *OrderQueries) GetOrderAuditLogs(ctx context.Context, id int64) ([]*AuditLog, error) {
	return s.store.OrderAuditLogs(ctx, id)
}
