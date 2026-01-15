package controller

import "github.com/gofiber/fiber/v2"

type PaymentController interface {
	Create(c *fiber.Ctx) error
	MarkAsSuccess(c *fiber.Ctx) error
	MarkAsFailed(c *fiber.Ctx) error
	FindById(c *fiber.Ctx) error
}
