package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"order-service/controller"
	"order-service/models/domain"
	"order-service/models/web"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrderService is a mock implementation of OrderService
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) Create(ctx context.Context, request web.OrderCreateRequest) (domain.Order, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(domain.Order), args.Error(1)
}

func (m *MockOrderService) Update(ctx context.Context, request web.OrderUpdateRequest) (domain.Order, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(domain.Order), args.Error(1)
}

func (m *MockOrderService) Delete(ctx context.Context, orderId string) error {
	args := m.Called(ctx, orderId)
	return args.Error(0)
}

func (m *MockOrderService) FindById(ctx context.Context, orderId string) (domain.Order, error) {
	args := m.Called(ctx, orderId)
	if args.Get(0) == nil {
		return domain.Order{}, args.Error(1)
	}
	return args.Get(0).(domain.Order), args.Error(1)
}

func (m *MockOrderService) FindAll(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return []domain.Order{}, args.Error(1)
	}
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *MockOrderService) ProcessPaymentCallback(ctx context.Context, request web.PaymentCallbackRequest) (domain.Order, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(domain.Order), args.Error(1)
}

// SUCCESS CONDITION TESTS

// Test Create endpoint
func TestOrderControllerCreate(t *testing.T) {
	mockService := new(MockOrderService)
	ctrl := controller.NewOrderController(mockService)

	app := fiber.New()
	app.Post("/orders", ctrl.Create)

	orderId := uuid.New()
	request := web.OrderCreateRequest{
		ItemName: "Test Item",
		Quantity: 2,
		Price:    1000,
	}

	expectedOrder := domain.Order{
		ID:          orderId,
		ItemName:    "Test Item",
		Quantity:    2,
		Price:       1000,
		TotalAmount: 2000,
		Status:      "pending",
	}

	mockService.On("Create", mock.Anything, request).Return(expectedOrder, nil)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

// Test FindById endpoint
func TestOrderControllerFindById(t *testing.T) {
	mockService := new(MockOrderService)
	ctrl := controller.NewOrderController(mockService)

	app := fiber.New()
	app.Get("/orders/:orderId", ctrl.FindById)

	orderId := uuid.New()
	expectedOrder := domain.Order{
		ID:          orderId,
		ItemName:    "Test Item",
		Quantity:    1,
		Price:       1000,
		TotalAmount: 1000,
		Status:      "pending",
	}

	mockService.On("FindById", mock.Anything, orderId.String()).Return(expectedOrder, nil)

	req := httptest.NewRequest(http.MethodGet, "/orders/"+orderId.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

// Test FindAll endpoint
func TestOrderControllerFindAll(t *testing.T) {
	mockService := new(MockOrderService)
	ctrl := controller.NewOrderController(mockService)

	app := fiber.New()
	app.Get("/orders", ctrl.FindAll)

	expectedOrders := []domain.Order{
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
			Status:      "pending",
		},
	}

	mockService.On("FindAll", mock.Anything).Return(expectedOrders, nil)

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

// Test Update endpoint
func TestOrderControllerUpdate(t *testing.T) {
	mockService := new(MockOrderService)
	ctrl := controller.NewOrderController(mockService)

	app := fiber.New()
	app.Put("/orders/:orderId", ctrl.Update)

	orderId := uuid.New()
	request := web.OrderUpdateRequest{
		ItemName: "Updated Item",
		Quantity: 3,
		Price:    2000,
	}

	updatedOrder := domain.Order{
		ID:          orderId,
		ItemName:    "Updated Item",
		Quantity:    3,
		Price:       2000,
		TotalAmount: 6000,
		Status:      "pending",
	}

	mockService.On("Update", mock.Anything, mock.MatchedBy(func(r web.OrderUpdateRequest) bool {
		return r.ItemName == "Updated Item"
	})).Return(updatedOrder, nil)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPut, "/orders/"+orderId.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

// Test Delete endpoint
func TestOrderControllerDelete(t *testing.T) {
	mockService := new(MockOrderService)
	ctrl := controller.NewOrderController(mockService)

	app := fiber.New()
	app.Delete("/orders/:orderId", ctrl.Delete)

	orderId := uuid.New()

	mockService.On("Delete", mock.Anything, orderId.String()).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/orders/"+orderId.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mockService.AssertExpectations(t)
}

// ERROR CONDITION TESTS

// Test Create with validation error (missing fields)
func TestOrderController_Create_ValidationError(t *testing.T) {
	mockService := new(MockOrderService)
	ctrl := controller.NewOrderController(mockService)

	app := fiber.New()
	app.Post("/orders", func(c *fiber.Ctx) error {
		return ctrl.Create(c)
	})

	requestBody := web.OrderCreateRequest{
		ItemName: "",
		Quantity: 1,
		Price:    1000,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	mockService.AssertNotCalled(t, "Create")
}

// Test Create when service fails
func TestOrderControllerCreateServiceError(t *testing.T) {
	mockService := new(MockOrderService)
	ctrl := controller.NewOrderController(mockService)

	app := fiber.New()
	app.Post("/orders", ctrl.Create)

	request := web.OrderCreateRequest{
		ItemName: "Test Item",
		Quantity: 2,
		Price:    1000,
	}

	// Mock service Create to return error
	mockService.On("Create", mock.Anything, request).Return(domain.Order{}, assert.AnError)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// Test Create with invalid JSON
func TestOrderControllerCreateInvalidJSON(t *testing.T) {
	mockService := new(MockOrderService)
	ctrl := controller.NewOrderController(mockService)

	app := fiber.New()
	app.Post("/orders", ctrl.Create)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// Test FindById when service fails
func TestOrderControllerFindByIdServiceError(t *testing.T) {
	mockService := new(MockOrderService)
	ctrl := controller.NewOrderController(mockService)

	app := fiber.New()
	app.Get("/orders/:orderId", ctrl.FindById)

	orderId := uuid.New()

	// Mock service to return error
	mockService.On("FindById", mock.Anything, orderId.String()).Return(domain.Order{}, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/orders/"+orderId.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// Test FindById with invalid UUID
func TestOrderControllerFindByIdInvalidUUID(t *testing.T) {
	mockService := new(MockOrderService)
	ctrl := controller.NewOrderController(mockService)

	app := fiber.New()
	app.Get("/orders/:orderId", ctrl.FindById)

	req := httptest.NewRequest(http.MethodGet, "/orders/invalid-uuid", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// Test Update with invalid UUID
func TestOrderControllerUpdateInvalidUUID(t *testing.T) {
	mockService := new(MockOrderService)
	ctrl := controller.NewOrderController(mockService)

	app := fiber.New()
	app.Put("/orders/:orderId", ctrl.Update)

	request := web.OrderUpdateRequest{
		ItemName: "Updated",
		Quantity: 1,
		Price:    1000,
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPut, "/orders/invalid-uuid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// Test Update when service fails
func TestOrderControllerUpdateServiceError(t *testing.T) {
	mockService := new(MockOrderService)
	ctrl := controller.NewOrderController(mockService)

	app := fiber.New()
	app.Put("/orders/:orderId", ctrl.Update)

	orderId := uuid.New()
	request := web.OrderUpdateRequest{
		ItemName: "Updated",
		Quantity: 1,
		Price:    1000,
	}

	// Mock service to return error
	mockService.On("Update", mock.Anything, mock.Anything).Return(domain.Order{}, assert.AnError)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPut, "/orders/"+orderId.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// Test Delete with invalid UUID
func TestOrderControllerDeleteInvalidUUID(t *testing.T) {
	mockService := new(MockOrderService)
	ctrl := controller.NewOrderController(mockService)

	app := fiber.New()
	app.Delete("/orders/:orderId", ctrl.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/orders/invalid-uuid", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// Test Delete when service fails
func TestOrderControllerDeleteServiceError(t *testing.T) {
	mockService := new(MockOrderService)
	ctrl := controller.NewOrderController(mockService)

	app := fiber.New()
	app.Delete("/orders/:orderId", ctrl.Delete)

	orderId := uuid.New()

	// Mock service to return error
	mockService.On("Delete", mock.Anything, orderId.String()).Return(assert.AnError)

	req := httptest.NewRequest(http.MethodDelete, "/orders/"+orderId.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// Test FindAll when service fails
func TestOrderControllerFindAllServiceError(t *testing.T) {
	mockService := new(MockOrderService)
	ctrl := controller.NewOrderController(mockService)

	app := fiber.New()
	app.Get("/orders", ctrl.FindAll)

	// Mock service to return error
	mockService.On("FindAll", mock.Anything).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
