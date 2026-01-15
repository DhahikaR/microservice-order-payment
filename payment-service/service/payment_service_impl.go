package service

import (
	"context"

	"errors"
	"fmt"

	"payment-service/helper"
	"payment-service/models/domain"
	"payment-service/models/web"
	"payment-service/repository"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentServiceImpl struct {
	PaymentRepository repository.PaymentRepository
	DB                *gorm.DB
	Validate          *validator.Validate
}

func NewPaymentService(paymentRepository repository.PaymentRepository, DB *gorm.DB, validate *validator.Validate) PaymentService {
	return &PaymentServiceImpl{
		PaymentRepository: paymentRepository,
		DB:                DB,
		Validate:          validate,
	}
}

func (service *PaymentServiceImpl) Create(ctx context.Context, request web.PaymentCreateRequest) (domain.Payment, error) {
	if err := service.Validate.Struct(request); err != nil {
		return domain.Payment{}, err
	}

	// Fetch order and validate amount
	orderTotalAmount, err := service.fetchOrderAmount(ctx, request.OrderID)
	if err != nil {
		return domain.Payment{}, err
	}

	// Validate that payment amount matches order total amount
	if request.Amount != orderTotalAmount {
		return domain.Payment{}, fmt.Errorf("payment amount %d does not match order total amount %d", request.Amount, orderTotalAmount)
	}

	// Validate if payment already exists for the order
	existingPayment, err := service.PaymentRepository.FindOrderById(ctx, service.DB, request.OrderID.String())
	if err == nil && existingPayment.ID != uuid.Nil {
		fmt.Printf("Payment already exists for order %s: returning payment %s",
			request.OrderID.String(), existingPayment.ID.String())
		return existingPayment, nil
	}

	tx := service.DB.Begin()
	defer helper.CommitOrRollback(tx)

	payment := domain.Payment{
		ID:       uuid.New(),
		OrderID:  request.OrderID,
		Amount:   request.Amount,
		Provider: request.Provider,
		Status:   "pending",
	}

	saved, err := service.PaymentRepository.Save(ctx, tx, payment)
	if err != nil {
		return domain.Payment{}, err
	}

	return saved, nil
}

func (service *PaymentServiceImpl) MarkAsSuccess(ctx context.Context, paymentId string) (domain.Payment, error) {
	tx := service.DB.Begin()
	defer helper.CommitOrRollback(tx)

	payment, err := service.PaymentRepository.FindById(ctx, tx, paymentId)
	if err != nil {
		return domain.Payment{}, err
	}

	if payment.Status != "pending" {
		return domain.Payment{}, errors.New("payment already finalized")
	}

	now := time.Now()
	payment.Status = "success"
	payment.PaidAt = &now

	updated, err := service.PaymentRepository.UpdateStatus(ctx, tx, payment)
	if err != nil {
		return domain.Payment{}, err
	}

	callbackPayload := web.PaymentCallbackRequest{
		OrderID:       updated.OrderID,
		PaymentID:     updated.ID,
		PaymentStatus: "success",
	}

	if err := SendPaymentCallback(ctx, service.getCallbackURL(), callbackPayload); err != nil {
		fmt.Printf("Warning: callback to order service failed: %v", err)
	}

	return updated, nil
}

func (service *PaymentServiceImpl) MarkAsFailed(ctx context.Context, paymentId string) (domain.Payment, error) {
	tx := service.DB.Begin()
	defer helper.CommitOrRollback(tx)

	payment, err := service.PaymentRepository.FindById(ctx, tx, paymentId)
	if err != nil {
		return domain.Payment{}, err
	}

	if payment.Status != "pending" {
		return domain.Payment{}, errors.New("payment already finalized")
	}

	payment.Status = "failed"
	payment.PaidAt = nil

	updated, err := service.PaymentRepository.UpdateStatus(ctx, tx, payment)
	if err != nil {
		return domain.Payment{}, err
	}

	callbackPayload := web.PaymentCallbackRequest{
		OrderID:       updated.OrderID,
		PaymentID:     updated.ID,
		PaymentStatus: "failed",
	}

	if err := SendPaymentCallback(ctx, service.getCallbackURL(), callbackPayload); err != nil {
		fmt.Printf("Warning: failed payment callback to order service failed: %v", err)
	}

	return updated, nil
}

func (service *PaymentServiceImpl) FindById(ctx context.Context, paymentId string) (domain.Payment, error) {
	tx := service.DB.Begin()
	defer helper.CommitOrRollback(tx)

	result, err := service.PaymentRepository.FindById(ctx, tx, paymentId)
	if err != nil {
		return domain.Payment{}, err
	}

	return result, nil
}
