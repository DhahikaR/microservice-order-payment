package repository

import (
	"context"
	"order-service/models/domain"
	"time"

	"gorm.io/gorm"
)

type OrderRepositoryImpl struct {
	DB *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &OrderRepositoryImpl{
		DB: db,
	}
}

func (repository *OrderRepositoryImpl) Save(ctx context.Context, tx *gorm.DB, order domain.Order) (domain.Order, error) {
	err := tx.WithContext(ctx).Create(&order).Error
	return order, err
}

func (repository *OrderRepositoryImpl) Update(ctx context.Context, tx *gorm.DB, order domain.Order) (domain.Order, error) {
	err := tx.WithContext(ctx).Model(domain.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
		"item_name":    order.ItemName,
		"quantity":     order.Quantity,
		"price":        order.Price,
		"total_amount": order.TotalAmount,
		"status":       order.Status,
		"payment_id":   order.PaymentID,
		"updated_at":   time.Now(),
	}).Error
	return order, err
}

func (repository *OrderRepositoryImpl) Delete(ctx context.Context, tx *gorm.DB, orderId string) error {
	return tx.WithContext(ctx).Model(&domain.Order{}).Where("id = ?", orderId).Update("deleted_at", gorm.DeletedAt{Valid: true}).Error
}

func (repository *OrderRepositoryImpl) FindById(ctx context.Context, tx *gorm.DB, orderId string) (domain.Order, error) {
	var order domain.Order
	result := tx.WithContext(ctx).Where("id = ?", orderId).First(&order)

	if result.Error != nil {
		return order, result.Error
	}

	return order, result.Error
}

func (repository *OrderRepositoryImpl) FindByAll(ctx context.Context, tx *gorm.DB) ([]domain.Order, error) {
	var orders []domain.Order
	err := tx.WithContext(ctx).Find(&orders).Error

	return orders, err
}
