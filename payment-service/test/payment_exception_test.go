package test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"payment-service/exception"
	"payment-service/models/web"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupApp() *fiber.App {
	return fiber.New(fiber.Config{ErrorHandler: exception.NewErrorHandler})
}

func decodeResponse(t *testing.T, response *http.Response) web.WebResponse {
	var webResponse web.WebResponse
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(&webResponse); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return webResponse
}

// TestNotFoundError tests the NotFoundError type
func TestErrorHandler_NotFoundError(t *testing.T) {
	app := setupApp()

	app.Get("/notfound", func(c *fiber.Ctx) error {
		return exception.NotFoundError{Message: "user not found"}
	})

	req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
	resp, _ := app.Test(req, -1)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	writer := decodeResponse(t, resp)
	assert.Equal(t, http.StatusNotFound, writer.Code)
	assert.Equal(t, "NOT FOUND", writer.Status)

	if message, ok := writer.Data.(string); ok {
		assert.Equal(t, "user not found", message)
	} else {
		t.Fatalf("expected Data to be string, got %T", writer.Data)
	}
}

func TestErrorHandler_FiberError(t *testing.T) {
	app := setupApp()

	app.Get("/fe", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	})

	req := httptest.NewRequest(http.MethodGet, "/fe", nil)
	resp, _ := app.Test(req, -1)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	writer := decodeResponse(t, resp)
	assert.Equal(t, http.StatusNotFound, writer.Code)
	assert.Equal(t, "NOT FOUND", writer.Status)
}

func TestErrorHandler_InternalServerError(t *testing.T) {
	app := setupApp()

	app.Get("/ise", func(c *fiber.Ctx) error {
		return errors.New("error")
	})

	req := httptest.NewRequest(http.MethodGet, "/ise", nil)
	resp, _ := app.Test(req, -1)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	writer := decodeResponse(t, resp)
	assert.Equal(t, http.StatusInternalServerError, writer.Code)
	assert.Equal(t, "INTERNAL SERVER ERROR", writer.Status)
}
