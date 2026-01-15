package web

import (
	"time"

	"github.com/google/uuid"
)

type PaymentResponse struct {
	ID        uuid.UUID  `json:"id"`
	OrderID   uuid.UUID  `json:"order_id"`
	Amount    int64      `json:"amount"`
	Status    string     `json:"status"`
	Provider  string     `json:"provider"`
	PaidAt    *time.Time `json:"paid_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
