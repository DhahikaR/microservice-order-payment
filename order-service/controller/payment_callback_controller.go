package controller

import (
	"order-service/helper"
	"order-service/models/web"
	"order-service/service"

	"github.com/gofiber/fiber/v2"
)

type PaymentCallbackController struct {
	orderService service.OrderService
}

func NewPaymentCallbackController(orderService service.OrderService) *PaymentCallbackController {
	return &PaymentCallbackController{orderService: orderService}
}

func (controller *PaymentCallbackController) Handle(c *fiber.Ctx) error {
	request := web.PaymentCallbackRequest{}

	if err := helper.ReadFromRequestBody(c, &request); err != nil {
		return helper.BadRequest(c, "invalid payload")
	}

	result, err := controller.orderService.ProcessPaymentCallback(c.Context(), request)

	if err != nil {
		return helper.BadRequest(c, err.Error())
	}

	return helper.ResponseSuccess(c, result)
}
