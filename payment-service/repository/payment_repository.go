package repository

import (
	"context"
	"payment-service/models/domain"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	Save(ctx context.Context, tx *gorm.DB, payment domain.Payment) (domain.Payment, error)
	FindById(ctx context.Context, tx *gorm.DB, paymentId string) (domain.Payment, error)
	UpdateStatus(ctx context.Context, tx *gorm.DB, payment domain.Payment) (domain.Payment, error)
	FindOrderById(ctx context.Context, tx *gorm.DB, orderId string) (domain.Payment, error)
}
