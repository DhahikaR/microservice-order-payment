package web

import "github.com/google/uuid"

type PaymentCallbackRequest struct {
	OrderID       uuid.UUID `json:"order_id" validate:"required"`
	PaymentID     uuid.UUID `json:"payment_id" validate:"required"`
	PaymentStatus string    `json:"payment_status" validate:"required,oneof=success failed"`
}
