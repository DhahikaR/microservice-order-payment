package routes

import (
	"payment-service/controller"

	"github.com/gofiber/fiber/v2"
)

func PaymentRoutes(app *fiber.App, paymentController controller.PaymentController) {
	payment := app.Group("/payments")

	payment.Post("/", paymentController.Create)
	payment.Get("/:paymentId", paymentController.FindById)
	payment.Put("/success/:paymentId", paymentController.MarkAsSuccess)
	payment.Put("/failed/:paymentId", paymentController.MarkAsFailed)
}
