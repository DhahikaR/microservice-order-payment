package controller

import (
	"payment-service/helper"
	"payment-service/models/web"
	"payment-service/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PaymentControllerImpl struct {
	paymentService service.PaymentService
}

func NewPaymentController(paymentService service.PaymentService) PaymentController {
	return &PaymentControllerImpl{
		paymentService: paymentService,
	}
}

func (controller *PaymentControllerImpl) Create(c *fiber.Ctx) error {
	request := web.PaymentCreateRequest{}
	if err := helper.ReadFromRequestBody(c, &request); err != nil {
		return helper.BadRequest(c, err.Error())
	}

	payment, err := controller.paymentService.Create(c.Context(), request)
	if err != nil {
		return helper.BadRequest(c, err.Error())
	}

	updated, err := controller.paymentService.MarkAsSuccess(c.Context(), payment.ID.String())
	if err != nil {
		return helper.InternalServerError(c, err.Error())
	}

	return helper.ResponseSuccess(c, updated)
}

func (controller *PaymentControllerImpl) FindById(c *fiber.Ctx) error {
	paymentId := c.Params("paymentId")

	if _, err := uuid.Parse(paymentId); err != nil {
		return helper.BadRequest(c, "invalid payment id")
	}

	result, err := controller.paymentService.FindById(c.Context(), paymentId)
	if err != nil {
		return helper.NotFound(c, "payment not found")
	}

	return helper.ResponseSuccess(c, result)
}

func (controller *PaymentControllerImpl) MarkAsSuccess(c *fiber.Ctx) error {
	paymentId := c.Params("paymentId")

	if _, err := uuid.Parse(paymentId); err != nil {
		return helper.BadRequest(c, "invalid payment id")
	}

	result, err := controller.paymentService.MarkAsSuccess(c.Context(), paymentId)
	if err != nil {
		return helper.BadRequest(c, err.Error())
	}

	return helper.ResponseSuccess(c, result)
}

func (controller *PaymentControllerImpl) MarkAsFailed(c *fiber.Ctx) error {
	paymentId := c.Params("paymentId")

	if _, err := uuid.Parse(paymentId); err != nil {
		return helper.BadRequest(c, "invalid payment id")
	}

	result, err := controller.paymentService.MarkAsFailed(c.Context(), paymentId)
	if err != nil {
		return helper.BadRequest(c, err.Error())
	}

	return helper.ResponseSuccess(c, result)
}
