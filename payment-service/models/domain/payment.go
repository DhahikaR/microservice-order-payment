package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Payment struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID   uuid.UUID      `gorm:"type:uuid;not null" json:"order_id"`
	Amount    int64          `json:"amount"`
	Status    string         `gorm:"type:varchar(50);default:'pending'" json:"status"`
	Provider  string         `json:"provider"`
	PaidAt    *time.Time     `json:"paid_at"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
