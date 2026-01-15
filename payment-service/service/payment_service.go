package service

import (
	"context"
	"payment-service/models/domain"
	"payment-service/models/web"
)

type PaymentService interface {
	Create(ctx context.Context, request web.PaymentCreateRequest) (domain.Payment, error)
	MarkAsSuccess(ctx context.Context, paymentId string) (domain.Payment, error)
	MarkAsFailed(ctx context.Context, paymentId string) (domain.Payment, error)
	FindById(ctx context.Context, paymentId string) (domain.Payment, error)
}
