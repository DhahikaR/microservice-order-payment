package helper

import "github.com/gofiber/fiber/v2"

func ReadFromRequestBody(c *fiber.Ctx, result interface{}) error {
	err := c.BodyParser(result)
	if err != nil {
		return err
	}

	return nil
}
