package repository

import (
	"context"
	"order-service/models/domain"

	"gorm.io/gorm"
)

type OrderRepository interface {
	Save(ctx context.Context, tx *gorm.DB, order domain.Order) (domain.Order, error)
	Update(ctx context.Context, tx *gorm.DB, order domain.Order) (domain.Order, error)
	Delete(ctx context.Context, tx *gorm.DB, orderId string) error
	FindById(ctx context.Context, tx *gorm.DB, orderId string) (domain.Order, error)
	FindByAll(ctx context.Context, tx *gorm.DB) ([]domain.Order, error)
}
