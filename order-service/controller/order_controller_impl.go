package controller

import (
	"order-service/helper"
	"order-service/models/web"
	"order-service/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type OrderControllerImpl struct {
	orderService service.OrderService
}

func NewOrderController(orderService service.OrderService) OrderController {
	return &OrderControllerImpl{
		orderService: orderService,
	}
}

func (controller *OrderControllerImpl) Create(c *fiber.Ctx) error {
	request := web.OrderCreateRequest{}
	if err := helper.ReadFromRequestBody(c, &request); err != nil {
		return helper.BadRequest(c, err.Error())
	}

	// Basic validation at controller level to avoid calling service with invalid input
	if request.ItemName == "" {
		return helper.BadRequest(c, "item name required")
	}

	if request.Quantity <= 0 {
		return helper.BadRequest(c, "quantity must be greater than 0")
	}

	if request.Price <= 0 {
		return helper.BadRequest(c, "price must be greater than 0")
	}

	order, err := controller.orderService.Create(c.Context(), request)
	if err != nil {
		return helper.InternalServerError(c, err.Error())
	}

	response := helper.ToOrderResponse(order)

	return helper.ResponseSuccess(c, response)
}

func (controller *OrderControllerImpl) Update(c *fiber.Ctx) error {
	request := web.OrderUpdateRequest{}
	if err := helper.ReadFromRequestBody(c, &request); err != nil {
		return helper.BadRequest(c, err.Error())
	}

	orderId := c.Params("orderId")

	if _, err := uuid.Parse(orderId); err != nil {
		return helper.BadRequest(c, "invalid UUID")
	}

	request.ID = uuid.MustParse(orderId)

	order, err := controller.orderService.Update(c.Context(), request)
	if err != nil {
		return helper.BadRequest(c, err.Error())
	}

	return helper.ResponseSuccess(c, helper.ToOrderResponse(order))
}

func (controller *OrderControllerImpl) Delete(c *fiber.Ctx) error {
	orderId := c.Params("orderId")
	if _, err := uuid.Parse(orderId); err != nil {
		return helper.BadRequest(c, "invalid UUID")
	}

	err := controller.orderService.Delete(c.Context(), orderId)
	if err != nil {
		return helper.BadRequest(c, "invalid order id")
	}

	return c.JSON(fiber.Map{
		"message": "order deleted",
		"id":      orderId,
	})
}

func (controller *OrderControllerImpl) FindById(c *fiber.Ctx) error {
	orderId := c.Params("orderId")
	if _, err := uuid.Parse(orderId); err != nil {
		return helper.BadRequest(c, "invalid UUID")
	}

	order, err := controller.orderService.FindById(c.Context(), orderId)
	if err != nil {
		return helper.BadRequest(c, err.Error())
	}

	response := helper.ToOrderResponse(order)

	return helper.ResponseSuccess(c, response)
}

func (controller *OrderControllerImpl) FindAll(c *fiber.Ctx) error {
	orders, err := controller.orderService.FindAll(c.Context())
	if err != nil {
		return helper.InternalServerError(c, "internal server error")
	}

	response := helper.ToOrderResponses(orders)

	return c.JSON(fiber.Map{"data": response})
}
