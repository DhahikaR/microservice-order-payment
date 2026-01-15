package main

import (
	"log"
	"order-service/config"
	"order-service/controller"
	"order-service/exception"
	"order-service/models/domain"
	"order-service/repository"
	"order-service/routes"
	"order-service/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: exception.NewErrorHandler,
	})

	db := config.NewDB()
	db.AutoMigrate(&domain.Order{})
	validate := validator.New()

	orderRepository := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepository, db, validate)
	orderController := controller.NewOrderController(orderService)
	paymentCallbackController := controller.NewPaymentCallbackController(orderService)

	routes.OrderRoutes(app, orderController)
	routes.PaymentCallbackRoutes(app, *paymentCallbackController)

	app.Listen(":3000")
}
