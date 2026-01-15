package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"payment-service/models/domain"
	"payment-service/models/web"
	"payment-service/service"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Reuse MockPaymentRepository from other tests in this folder.

// TestFetchOrderAmountSuccess validates that when order-service returns a valid
// total_amount the payment Create flow succeeds (indirectly exercising fetchOrderAmount).
func TestFetchOrderAmountSuccess(t *testing.T) {
	orderTotal := int64(2500)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 200,
			"data": map[string]interface{}{"total_amount": orderTotal},
		})
	}))
	defer srv.Close()

	os.Setenv("ORDER_SERVICE_URL", srv.URL)

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.Payment{})

	mockRepo := new(MockPaymentRepository)
	validate := validator.New()
	svc := service.NewPaymentService(mockRepo, db, validate)

	orderId := uuid.New()
	req := web.PaymentCreateRequest{OrderID: orderId, Amount: orderTotal, Provider: "x"}

	// simulate no existing payment
	mockRepo.On("FindOrderById", mock.Anything, mock.Anything, orderId.String()).Return(domain.Payment{}, assert.AnError)

	// expect Save to be called and return the payment
	expected := domain.Payment{ID: uuid.New(), OrderID: orderId, Amount: orderTotal, Provider: "x", Status: "pending"}
	mockRepo.On("Save", mock.Anything, mock.Anything, mock.MatchedBy(func(p domain.Payment) bool { return p.OrderID == orderId })).Return(expected, nil)

	got, err := svc.Create(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, expected.OrderID, got.OrderID)
	assert.Equal(t, expected.Amount, got.Amount)
	mockRepo.AssertExpectations(t)
}

// TestFetchOrderAmountNotFound ensures a 404 from order-service surfaces as an error
func TestFetchOrderAmountNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	os.Setenv("ORDER_SERVICE_URL", srv.URL)

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.Payment{})

	mockRepo := new(MockPaymentRepository)
	validate := validator.New()
	svc := service.NewPaymentService(mockRepo, db, validate)

	orderId := uuid.New()
	req := web.PaymentCreateRequest{OrderID: orderId, Amount: 1000, Provider: "x"}

	_, err := svc.Create(context.Background(), req)
	assert.Error(t, err)
}

// TestFetchOrderAmountInvalidJSON ensures that invalid JSON from order-service errors
func TestFetchOrderAmountInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// send invalid json
		w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	os.Setenv("ORDER_SERVICE_URL", srv.URL)

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.Payment{})

	mockRepo := new(MockPaymentRepository)
	validate := validator.New()
	svc := service.NewPaymentService(mockRepo, db, validate)

	orderId := uuid.New()
	req := web.PaymentCreateRequest{OrderID: orderId, Amount: 1000, Provider: "x"}

	_, err := svc.Create(context.Background(), req)
	assert.Error(t, err)
}

// TestFetchOrderAmountServerError ensures that non-200/404 responses result in error
func TestFetchOrderAmountServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	os.Setenv("ORDER_SERVICE_URL", srv.URL)

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.Payment{})

	mockRepo := new(MockPaymentRepository)
	validate := validator.New()
	svc := service.NewPaymentService(mockRepo, db, validate)

	orderId := uuid.New()
	req := web.PaymentCreateRequest{OrderID: orderId, Amount: 1000, Provider: "x"}

	_, err := svc.Create(context.Background(), req)
	assert.Error(t, err)
}
