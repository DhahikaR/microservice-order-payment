package routes

import (
	"order-service/controller"

	"github.com/gofiber/fiber/v2"
)

func OrderRoutes(app *fiber.App, orderController controller.OrderController) {
	order := app.Group("/orders")

	order.Get("/", orderController.FindAll)
	order.Get("/:orderId", orderController.FindById)
	order.Post("/", orderController.Create)
	order.Put("/:orderId", orderController.Update)
	order.Delete("/:orderId", orderController.Delete)
}

func PaymentCallbackRoutes(app *fiber.App, callbackController controller.PaymentCallbackController) {
	app.Post("/internal/payment-callback", callbackController.Handle)
}
