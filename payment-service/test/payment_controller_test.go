package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"payment-service/controller"
	"payment-service/models/domain"
	"payment-service/models/web"

	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPaymentService is a mock for payment service
type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) Create(ctx context.Context, request web.PaymentCreateRequest) (domain.Payment, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(domain.Payment), args.Error(1)
}
func (m *MockPaymentService) MarkAsSuccess(ctx context.Context, paymentId string) (domain.Payment, error) {
	args := m.Called(ctx, paymentId)
	return args.Get(0).(domain.Payment), args.Error(1)
}
func (m *MockPaymentService) MarkAsFailed(ctx context.Context, paymentId string) (domain.Payment, error) {
	args := m.Called(ctx, paymentId)
	return args.Get(0).(domain.Payment), args.Error(1)
}
func (m *MockPaymentService) FindById(ctx context.Context, paymentId string) (domain.Payment, error) {
	args := m.Called(ctx, paymentId)
	if args.Get(0) == nil {
		return domain.Payment{}, args.Error(1)
	}
	return args.Get(0).(domain.Payment), args.Error(1)
}

// TestCreateSuccess tests controller Create happy path
func TestCreateSuccess(t *testing.T) {
	svc := new(MockPaymentService)
	ctrl := controller.NewPaymentController(svc)

	app := fiber.New()
	app.Post("/payments", ctrl.Create)

	req := web.PaymentCreateRequest{OrderID: uuid.New(), Amount: 1000, Provider: "x"}
	created := domain.Payment{ID: uuid.New(), OrderID: req.OrderID, Amount: req.Amount, Status: "success"}

	svc.On("Create", mock.Anything, req).Return(domain.Payment{ID: created.ID, OrderID: req.OrderID, Amount: req.Amount, Provider: req.Provider, Status: "pending"}, nil)
	svc.On("MarkAsSuccess", mock.Anything, mock.Anything).Return(created, nil)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(r)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	svc.AssertExpectations(t)
}

// TestFindByIdSuccess tests controller FindById happy path
func TestFindByIdSuccess(t *testing.T) {
	svc := new(MockPaymentService)
	ctrl := controller.NewPaymentController(svc)

	app := fiber.New()
	app.Get("/payments/:paymentId", ctrl.FindById)

	id := uuid.New()
	found := domain.Payment{ID: id, OrderID: uuid.New(), Amount: 1000, Status: "success"}

	svc.On("FindById", mock.Anything, id.String()).Return(found, nil)

	r := httptest.NewRequest(http.MethodGet, "/payments/"+id.String(), nil)
	resp, _ := app.Test(r)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	svc.AssertExpectations(t)
}

// TestMarkAsSuccess tests controller MarkAsSuccess happy path
func TestMarkAsSuccess(t *testing.T) {
	svc := new(MockPaymentService)
	ctrl := controller.NewPaymentController(svc)

	app := fiber.New()
	app.Post("/payments/:paymentId/success", ctrl.MarkAsSuccess)

	id := uuid.New()
	updated := domain.Payment{ID: id, OrderID: uuid.New(), Amount: 1000, Status: "success"}

	svc.On("MarkAsSuccess", mock.Anything, id.String()).Return(updated, nil)

	r := httptest.NewRequest(http.MethodPost, "/payments/"+id.String()+"/success", nil)
	resp, _ := app.Test(r)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	svc.AssertExpectations(t)
}

// Test Error Condition

// TestCreateServiceError tests controller Create when service returns error
func TestCreateServiceError(t *testing.T) {
	svc := new(MockPaymentService)
	ctrl := controller.NewPaymentController(svc)

	app := fiber.New()
	app.Post("/payments", ctrl.Create)

	req := web.PaymentCreateRequest{OrderID: uuid.New(), Amount: 1000, Provider: "x"}
	svc.On("Create", mock.Anything, req).Return(domain.Payment{}, assert.AnError)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(r)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestCreateInvalidJSON tests controller Create with invalid JSON
func TestCreateInvalidJSON(t *testing.T) {
	svc := new(MockPaymentService)
	ctrl := controller.NewPaymentController(svc)

	app := fiber.New()
	app.Post("/payments", ctrl.Create)

	r := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewReader([]byte("{invalid")))
	r.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(r)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestFindByIdNotFound tests FindById when service returns error
func TestFindByIdNotFound(t *testing.T) {
	svc := new(MockPaymentService)
	ctrl := controller.NewPaymentController(svc)

	app := fiber.New()
	app.Get("/payments/:paymentId", ctrl.FindById)

	id := uuid.New()
	svc.On("FindById", mock.Anything, id.String()).Return(domain.Payment{}, assert.AnError)

	r := httptest.NewRequest(http.MethodGet, "/payments/"+id.String(), nil)
	resp, _ := app.Test(r)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestMarkAsSuccessInvalidUUID tests invalid UUID handling
func TestMarkAsSuccessInvalidUUID(t *testing.T) {
	svc := new(MockPaymentService)
	ctrl := controller.NewPaymentController(svc)

	app := fiber.New()
	app.Post("/payments/:paymentId/success", ctrl.MarkAsSuccess)

	r := httptest.NewRequest(http.MethodPost, "/payments/invalid-uuid/success", nil)
	resp, _ := app.Test(r)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
