package service

import (
	"context"
	"errors"
	"order-service/helper"
	"order-service/models/domain"
	"order-service/models/web"
	"order-service/repository"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type OrderServiceImpl struct {
	OrderRepository repository.OrderRepository
	DB              *gorm.DB
	Validate        *validator.Validate
}

func NewOrderService(orderRepository repository.OrderRepository, DB *gorm.DB, validate *validator.Validate) OrderService {
	return &OrderServiceImpl{
		OrderRepository: orderRepository,
		DB:              DB,
		Validate:        validate,
	}
}

func (service *OrderServiceImpl) Create(ctx context.Context, request web.OrderCreateRequest) (domain.Order, error) {
	err := service.Validate.Struct(request)
	helper.PanicIfError(err)

	tx := service.DB.Begin()
	defer helper.CommitOrRollback(tx)

	order := domain.Order{
		ItemName:    request.ItemName,
		Quantity:    request.Quantity,
		Price:       request.Price,
		TotalAmount: request.Price * int64(request.Quantity),
		Status:      "pending",
	}

	created, err := service.OrderRepository.Save(ctx, tx, order)
	if err != nil {
		return domain.Order{}, err
	}

	return created, nil
}

func (service *OrderServiceImpl) Update(ctx context.Context, request web.OrderUpdateRequest) (domain.Order, error) {
	if err := service.Validate.Struct(request); err != nil {
		return domain.Order{}, err
	}

	tx := service.DB.Begin()
	defer helper.CommitOrRollback(tx)

	order, err := service.OrderRepository.FindById(ctx, tx, request.ID.String())
	if err != nil {
		return domain.Order{}, err
	}

	if order.Status == "paid" {
		return domain.Order{}, errors.New("paid order cannot be updated")
	}

	if request.Quantity <= 0 {
		return domain.Order{}, errors.New("quantity must be greater than 0")
	}

	if request.Price <= 0 {
		return domain.Order{}, errors.New("price must be greater than 0")
	}

	order.ItemName = request.ItemName
	order.Quantity = request.Quantity
	order.Price = request.Price
	order.TotalAmount = request.Price * int64(request.Quantity)

	updated, err := service.OrderRepository.Update(ctx, tx, order)
	if err != nil {
		return domain.Order{}, err
	}

	return updated, err
}

func (service *OrderServiceImpl) Delete(ctx context.Context, orderId string) error {
	tx := service.DB.Begin()
	defer helper.CommitOrRollback(tx)

	order, err := service.OrderRepository.FindById(ctx, tx, orderId)
	if err != nil {
		return err
	}

	if order.Status == "paid" {
		return errors.New("paid order cannot be deleted")
	}

	if err := service.OrderRepository.Delete(ctx, tx, orderId); err != nil {
		return err
	}

	return nil
}

func (service *OrderServiceImpl) FindById(ctx context.Context, orderId string) (domain.Order, error) {
	tx := service.DB.Begin()
	defer helper.CommitOrRollback(tx)

	order, err := service.OrderRepository.FindById(ctx, tx, orderId)
	if err != nil {
		return domain.Order{}, err
	}

	return order, err
}

func (service *OrderServiceImpl) FindAll(ctx context.Context) ([]domain.Order, error) {
	tx := service.DB.Begin()
	defer helper.CommitOrRollback(tx)

	orders, err := service.OrderRepository.FindByAll(ctx, tx)
	if err != nil {
		return []domain.Order{}, err
	}

	return orders, err
}

func (service *OrderServiceImpl) ProcessPaymentCallback(ctx context.Context, request web.PaymentCallbackRequest) (domain.Order, error) {
	tx := service.DB.Begin()
	defer helper.CommitOrRollback(tx)

	order, err := service.OrderRepository.FindById(ctx, tx, request.OrderID.String())
	if err != nil {
		return domain.Order{}, err
	}

	if request.PaymentStatus == "success" {
		order.Status = "paid"
	}

	return service.OrderRepository.Update(ctx, tx, order)
}
