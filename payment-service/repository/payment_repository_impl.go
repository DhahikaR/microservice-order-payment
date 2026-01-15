package repository

import (
	"context"
	"payment-service/models/domain"

	"gorm.io/gorm"
)

type PaymentRepositoryImpl struct {
	DB *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &PaymentRepositoryImpl{
		DB: db,
	}
}

func (repository *PaymentRepositoryImpl) Save(ctx context.Context, tx *gorm.DB, payment domain.Payment) (domain.Payment, error) {
	err := tx.WithContext(ctx).Create(&payment).Error

	return payment, err
}

func (repository *PaymentRepositoryImpl) FindById(ctx context.Context, tx *gorm.DB, paymentId string) (domain.Payment, error) {
	var payment domain.Payment
	err := tx.WithContext(ctx).Where("id = ?", paymentId).First(&payment).Error

	return payment, err
}

func (repository *PaymentRepositoryImpl) UpdateStatus(ctx context.Context, tx *gorm.DB, payment domain.Payment) (domain.Payment, error) {
	err := tx.WithContext(ctx).Model(&domain.Payment{}).Where("id = ?", payment.ID).Updates(map[string]interface{}{
		"status":  payment.Status,
		"paid_at": payment.PaidAt,
	}).Error

	return payment, err
}

func (repository *PaymentRepositoryImpl) FindOrderById(ctx context.Context, tx *gorm.DB, orderId string) (domain.Payment, error) {
	var payment domain.Payment
	err := tx.WithContext(ctx).Where("order_id = ?", orderId).First(&payment).Error

	return payment, err
}
