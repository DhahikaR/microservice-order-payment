package service

import (
	"context"
	"order-service/models/domain"
	"order-service/models/web"
)

type OrderService interface {
	Create(ctx context.Context, request web.OrderCreateRequest) (domain.Order, error)
	Update(ctx context.Context, request web.OrderUpdateRequest) (domain.Order, error)
	Delete(ctx context.Context, orderId string) error
	FindById(ctx context.Context, orderId string) (domain.Order, error)
	FindAll(ctx context.Context) ([]domain.Order, error)
	ProcessPaymentCallback(ctx context.Context, request web.PaymentCallbackRequest) (domain.Order, error)
}
