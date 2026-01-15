package exception

import (
	"payment-service/models/web"

	"github.com/gofiber/fiber/v2"
)

func NewErrorHandler(c *fiber.Ctx, err error) error {
	if notFound, ok := err.(NotFoundError); ok {
		return c.Status(fiber.StatusNotFound).JSON(web.WebResponse{
			Code:   fiber.StatusNotFound,
			Status: "NOT FOUND",
			Data:   notFound.Error(),
		})
	}

	if fiberError, ok := err.(*fiber.Error); ok {
		code := fiberError.Code
		if code == 0 {
			code = fiber.StatusInternalServerError
		}
		statusText := "ERROR"
		if code == fiber.StatusBadRequest {
			statusText = "BAD REQUEST"
		} else if code == fiber.StatusNotFound {
			statusText = "NOT FOUND"
		} else if code == fiber.StatusInternalServerError {
			statusText = "INTERNAL SERVICE ERROR"
		}
		return c.Status(code).JSON(web.WebResponse{
			Code:   code,
			Status: statusText,
			Data:   fiberError.Message,
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(web.WebResponse{
		Code:   fiber.StatusInternalServerError,
		Status: "INTERNAL SERVER ERROR",
		Data:   err.Error(),
	})
}
