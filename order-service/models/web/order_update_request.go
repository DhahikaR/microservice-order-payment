package web

import "github.com/google/uuid"

type OrderUpdateRequest struct {
	ID       uuid.UUID `validate:"required"`
	ItemName string    `validate:"required"`
	Quantity int       `validate:"required,gt=0"`
	Price    int64     `validate:"required,gt=0"`
}
