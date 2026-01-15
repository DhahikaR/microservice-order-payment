package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"order-service/helper"
	"order-service/models/domain"
	"order-service/models/web"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestPanicIfError tests the PanicIfError function
func TestPanicIfError(t *testing.T) {
	// Test with no error - should not panic
	assert.NotPanics(t, func() {
		helper.PanicIfError(nil)
	})

	// Test with error - should panic
	assert.Panics(t, func() {
		helper.PanicIfError(fiber.NewError(fiber.StatusBadRequest, "test error"))
	})
}

// TestErrorResponse tests the ErrorResponse function
func TestErrorResponse(t *testing.T) {
	err := fiber.NewError(fiber.StatusBadRequest, "test error")
	response := helper.ErrorResponse(err)

	assert.NotNil(t, response)
	assert.Equal(t, "test error", response["error"])
}

// TestToOrderResponse tests the ToOrderResponse function
func TestToOrderResponse(t *testing.T) {
	paymentId := uuid.New()
	order := domain.Order{
		ID:          uuid.New(),
		ItemName:    "Test Item",
		Quantity:    2,
		Price:       1000,
		TotalAmount: 2000,
		Status:      "pending",
		PaymentID:   &paymentId,
	}

	response := helper.ToOrderResponse(order)

	assert.NotNil(t, response)
	assert.Equal(t, order.ID, response.Id)
	assert.Equal(t, "Test Item", response.ItemName)
	assert.Equal(t, 2, response.Quantity)
	assert.Equal(t, int64(1000), response.Price)
	assert.Equal(t, int64(2000), response.TotalAmount)
	assert.Equal(t, "pending", response.Status)
}

// TestToOrderResponses tests the ToOrderResponses function
func TestToOrderResponses(t *testing.T) {
	orders := []domain.Order{
		{
			ID:          uuid.New(),
			ItemName:    "Item1",
			Quantity:    1,
			Price:       1000,
			TotalAmount: 1000,
			Status:      "pending",
		},
		{
			ID:          uuid.New(),
			ItemName:    "Item2",
			Quantity:    2,
			Price:       2000,
			TotalAmount: 4000,
			Status:      "paid",
		},
	}

	responses := helper.ToOrderResponses(orders)

	assert.Len(t, responses, 2)
	assert.Equal(t, "Item1", responses[0].ItemName)
	assert.Equal(t, "Item2", responses[1].ItemName)
}

// TestReadFromRequestBody
func TestResponseAndReadFromRequestBody(t *testing.T) {
	app := fiber.New()

	app.Post("/parse", func(c *fiber.Ctx) error {
		var req web.OrderCreateRequest
		if err := helper.ReadFromRequestBody(c, &req); err != nil {
			return helper.BadRequest(c, err.Error())
		}
		return helper.ResponseSuccess(c, req)
	})

	// valid JSON
	body, _ := json.Marshal(web.OrderCreateRequest{ItemName: "test", Quantity: 1, Price: 1000})
	req := httptest.NewRequest(http.MethodPost, "/parse", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// invalid JSON
	req2 := httptest.NewRequest(http.MethodPost, "/parse", bytes.NewReader([]byte(`{invalid}`)))
	req2.Header.Set("Content-Type", "application/json")
	resp2, _ := app.Test(req2, -1)

	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)
}

// TestToOrderResponseWithNilPaymentID tests conversion with nil PaymentID
func TestToOrderResponseWithNilPaymentID(t *testing.T) {
	order := domain.Order{
		ID:          uuid.New(),
		ItemName:    "Test Item",
		Quantity:    1,
		Price:       1000,
		TotalAmount: 1000,
		Status:      "pending",
		PaymentID:   nil,
	}

	response := helper.ToOrderResponse(order)

	assert.NotNil(t, response)
	assert.Equal(t, "Test Item", response.ItemName)
	assert.Equal(t, "pending", response.Status)
}

// TestToOrderResponseWithPaidStatus tests conversion with paid status
func TestToOrderResponseWithPaidStatus(t *testing.T) {
	paymentId := uuid.New()
	order := domain.Order{
		ID:          uuid.New(),
		ItemName:    "Paid Item",
		Quantity:    1,
		Price:       1000,
		TotalAmount: 1000,
		Status:      "paid",
		PaymentID:   &paymentId,
	}

	response := helper.ToOrderResponse(order)

	assert.NotNil(t, response)
	assert.Equal(t, "paid", response.Status)
}

// TestToOrderResponsesWithEmptyList tests converting empty order list
func TestToOrderResponsesWithEmptyList(t *testing.T) {
	orders := []domain.Order{}

	responses := helper.ToOrderResponses(orders)

	// Empty slice is valid, may be nil or empty depending on implementation
	assert.Len(t, responses, 0)
}

// TestToOrderResponsesWithMixedStatuses tests converting orders with different statuses
func TestToOrderResponsesWithMixedStatuses(t *testing.T) {
	orders := []domain.Order{
		{
			ID:          uuid.New(),
			ItemName:    "Pending Item",
			Quantity:    1,
			Price:       1000,
			TotalAmount: 1000,
			Status:      "pending",
		},
		{
			ID:          uuid.New(),
			ItemName:    "Paid Item",
			Quantity:    1,
			Price:       2000,
			TotalAmount: 2000,
			Status:      "paid",
		},
		{
			ID:          uuid.New(),
			ItemName:    "Failed Item",
			Quantity:    1,
			Price:       3000,
			TotalAmount: 3000,
			Status:      "failed",
		},
	}

	responses := helper.ToOrderResponses(orders)

	assert.Len(t, responses, 3)
	assert.Equal(t, "pending", responses[0].Status)
	assert.Equal(t, "paid", responses[1].Status)
	assert.Equal(t, "failed", responses[2].Status)
}

// TestErrorResponseWithDifferentErrors tests error response with various errors
func TestErrorResponseWithDifferentErrors(t *testing.T) {
	errors := []error{
		fiber.NewError(fiber.StatusBadRequest, "Bad Request"),
		fiber.NewError(fiber.StatusNotFound, "Not Found"),
		fiber.NewError(fiber.StatusInternalServerError, "Internal Error"),
	}

	for _, err := range errors {
		response := helper.ErrorResponse(err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response["error"])
	}
}

// TestPanicIfErrorWithDifferentErrors tests panic with various error types
func TestPanicIfErrorWithDifferentErrors(t *testing.T) {
	testCases := []struct {
		name        string
		err         error
		shouldPanic bool
	}{
		{"with nil error", nil, false},
		{"with fiber error", fiber.NewError(400, "bad request"), true},
		{"with another error", assert.AnError, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				assert.Panics(t, func() {
					helper.PanicIfError(tc.err)
				})
			} else {
				assert.NotPanics(t, func() {
					helper.PanicIfError(tc.err)
				})
			}
		})
	}
}
