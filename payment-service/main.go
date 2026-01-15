package main

import (
	"log"
	"payment-service/config"
	"payment-service/controller"
	"payment-service/exception"
	"payment-service/models/domain"
	"payment-service/repository"
	"payment-service/routes"
	"payment-service/service"

	"github.com/go-playground/validator"
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
	db.AutoMigrate(&domain.Payment{})
	validate := validator.New()

	paymentRepository := repository.NewPaymentRepository(db)
	paymentService := service.NewPaymentService(paymentRepository, db, validate)
	paymentController := controller.NewPaymentController(paymentService)

	routes.PaymentRoutes(app, paymentController)

	app.Listen(":3000")
}
