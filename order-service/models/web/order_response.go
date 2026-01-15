package web

import (
	"time"

	"github.com/google/uuid"
)

type OrderResponse struct {
	Id          uuid.UUID `json:"id"`
	ItemName    string    `json:"item_name"`
	Quantity    int       `json:"quantity"`
	Price       int64     `json:"price"`
	TotalAmount int64     `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
