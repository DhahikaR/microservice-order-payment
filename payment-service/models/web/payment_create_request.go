package web

import "github.com/google/uuid"

type PaymentCreateRequest struct {
	OrderID  uuid.UUID `json:"order_id" validate:"required"`
	Amount   int64     `json:"amount" validate:"required"`
	Provider string    `json:"provider" validate:"required"`
}
