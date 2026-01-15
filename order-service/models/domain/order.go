package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ItemName    string         `json:"item_name"`
	Quantity    int            `json:"quantity"`
	Price       int64          `json:"price"`
	TotalAmount int64          `json:"total_amount"`
	Status      string         `gorm:"type:varchar(50);default:'pending'" json:"status"`
	PaymentID   *uuid.UUID     `gorm:"type:uuid" json:"payment_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
