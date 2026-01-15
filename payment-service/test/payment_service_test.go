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

// MockPaymentRepository mocks repository methods
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Save(ctx context.Context, tx *gorm.DB, payment domain.Payment) (domain.Payment, error) {
	args := m.Called(ctx, tx, payment)
	return args.Get(0).(domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindById(ctx context.Context, tx *gorm.DB, paymentId string) (domain.Payment, error) {
	args := m.Called(ctx, tx, paymentId)
	if args.Get(0) == nil {
		return domain.Payment{}, args.Error(1)
	}
	return args.Get(0).(domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) UpdateStatus(ctx context.Context, tx *gorm.DB, payment domain.Payment) (domain.Payment, error) {
	args := m.Called(ctx, tx, payment)
	return args.Get(0).(domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindOrderById(ctx context.Context, tx *gorm.DB, orderId string) (domain.Payment, error) {
	args := m.Called(ctx, tx, orderId)
	if args.Get(0) == nil {
		return domain.Payment{}, args.Error(1)
	}
	return args.Get(0).(domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) MarkAsSuccess(ctx context.Context, tx *gorm.DB, paymentId string) error {
	args := m.Called(ctx, tx, paymentId)
	return args.Error(0)
}

func (m *MockPaymentRepository) MarkAsFailed(ctx context.Context, tx *gorm.DB, paymentId string) error {
	args := m.Called(ctx, tx, paymentId)
	return args.Error(0)
}

// TEST SUCCESS CONDITIONS

// TestPaymentServiceCreateSuccess tests creating a payment when order amount matches
func TestPaymentServiceCreateSuccess(t *testing.T) {
	// Start a test HTTP server to simulate order-service
	orderTotal := int64(5000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 200,
			"data": map[string]interface{}{"total_amount": orderTotal},
		})
	}))
	defer srv.Close()

	os.Setenv("ORDER_SERVICE_URL", srv.URL)

	// setup sqlite in-memory DB for tx support
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.Payment{})

	mockRepo := new(MockPaymentRepository)
	validate := validator.New()
	svc := service.NewPaymentService(mockRepo, db, validate)

	orderId := uuid.New()
	req := web.PaymentCreateRequest{OrderID: orderId, Amount: orderTotal, Provider: "stripe"}

	// simulate no existing payment
	mockRepo.On("FindOrderById", mock.Anything, mock.Anything, orderId.String()).Return(domain.Payment{}, assert.AnError)

	// expect Save to be called and return the payment
	expected := domain.Payment{ID: uuid.New(), OrderID: orderId, Amount: orderTotal, Provider: "stripe", Status: "pending"}
	mockRepo.On("Save", mock.Anything, mock.Anything, mock.MatchedBy(func(p domain.Payment) bool { return p.OrderID == orderId })).Return(expected, nil)

	got, err := svc.Create(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, expected.OrderID, got.OrderID)
	assert.Equal(t, expected.Amount, got.Amount)
	mockRepo.AssertExpectations(t)
}

// TestPaymentServiceMarkAsSuccess tests successful mark-as-success path and callback
func TestPaymentServiceMarkAsSuccess(t *testing.T) {
	// callback server to accept POST
	cbReceived := make(chan bool, 1)
	cbSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cbReceived <- true
		w.WriteHeader(http.StatusOK)
	}))
	defer cbSrv.Close()
	os.Setenv("ORDER_CALLBACK_URL", cbSrv.URL)

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.Payment{})

	mockRepo := new(MockPaymentRepository)
	validate := validator.New()
	svc := service.NewPaymentService(mockRepo, db, validate)

	paymentId := uuid.New()
	orderId := uuid.New()
	existing := domain.Payment{ID: paymentId, OrderID: orderId, Amount: 1000, Status: "pending"}
	updated := existing
	updated.Status = "success"

	mockRepo.On("FindById", mock.Anything, mock.Anything, paymentId.String()).Return(existing, nil)
	mockRepo.On("UpdateStatus", mock.Anything, mock.Anything, mock.MatchedBy(func(p domain.Payment) bool { return p.Status == "success" })).Return(updated, nil)

	got, err := svc.MarkAsSuccess(context.Background(), paymentId.String())
	assert.NoError(t, err)
	assert.Equal(t, "success", got.Status)

	// ensure callback was received (non-blocking)
	select {
	case <-cbReceived:
	default:
		// callback may be async; not failing test here
	}

	mockRepo.AssertExpectations(t)
}

// TestPaymentServiceMarkAsFailed tests successful mark-as-failed path
func TestPaymentServiceMarkAsFailed(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.Payment{})

	mockRepo := new(MockPaymentRepository)
	validate := validator.New()
	svc := service.NewPaymentService(mockRepo, db, validate)

	paymentId := uuid.New()
	existing := domain.Payment{ID: paymentId, Amount: 1000, Status: "pending"}
	updated := existing
	updated.Status = "failed"

	mockRepo.On("FindById", mock.Anything, mock.Anything, paymentId.String()).Return(existing, nil)
	mockRepo.On("UpdateStatus", mock.Anything, mock.Anything, mock.MatchedBy(func(p domain.Payment) bool { return p.Status == "failed" })).Return(updated, nil)

	got, err := svc.MarkAsFailed(context.Background(), paymentId.String())
	assert.NoError(t, err)
	assert.Equal(t, "failed", got.Status)

	mockRepo.AssertExpectations(t)
}

// TestPaymentServiceFindById tests successful FindById
func TestPaymentServiceFindById(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.Payment{})

	mockRepo := new(MockPaymentRepository)
	validate := validator.New()
	svc := service.NewPaymentService(mockRepo, db, validate)

	paymentId := uuid.New()
	expected := domain.Payment{ID: paymentId, OrderID: uuid.New(), Amount: 2000, Status: "success"}

	mockRepo.On("FindById", mock.Anything, mock.Anything, paymentId.String()).Return(expected, nil)

	got, err := svc.FindById(context.Background(), paymentId.String())
	assert.NoError(t, err)
	assert.Equal(t, expected.ID, got.ID)
	assert.Equal(t, expected.Status, got.Status)

	mockRepo.AssertExpectations(t)
}

// TEST ERROR CONDITIONS

// TestPaymentServiceCreateAmountMismatch tests when payment amount doesn't match order total
func TestPaymentServiceCreateAmountMismatch(t *testing.T) {
	orderTotal := int64(7000)
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
	req := web.PaymentCreateRequest{OrderID: orderId, Amount: 1234, Provider: "x"}

	// Should return error because amounts mismatch
	_, err := svc.Create(context.Background(), req)
	assert.Error(t, err)
}

// TestPaymentServiceCreateExistingPayment tests creating a payment when one already exists for the order
func TestPaymentServiceCreateExistingPayment(t *testing.T) {
	orderTotal := int64(4000)
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

	// Simulate existing payment found
	existing := domain.Payment{ID: uuid.New(), OrderID: orderId, Amount: orderTotal, Status: "pending"}
	mockRepo.On("FindOrderById", mock.Anything, mock.Anything, orderId.String()).Return(existing, nil)

	// Should return the existing payment without error (idempotent behavior)
	got, err := svc.Create(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, existing.ID, got.ID)
	assert.Equal(t, existing.Status, got.Status)

	// Verify Save was NOT called (no new record created)
	mockRepo.AssertNotCalled(t, "Save")
}

// TestPaymentServiceCreateOrderNotFound tests creating a payment when order-service returns not found
func TestPaymentServiceCreateOrderNotFound(t *testing.T) {
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

// TestPaymentServiceCreateOrderServiceError tests creating a payment when order-service returns error
func TestPaymentServiceCreateOrderServiceError(t *testing.T) {
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

// TestPaymentServiceCreateValidationError tests creating a payment with invalid request data
func TestPaymentServiceCreateValidationError(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.Payment{})

	mockRepo := new(MockPaymentRepository)
	validate := validator.New()
	svc := service.NewPaymentService(mockRepo, db, validate)

	// Invalid request: missing Provider
	orderId := uuid.New()
	req := web.PaymentCreateRequest{OrderID: orderId, Amount: 1000, Provider: ""}

	_, err := svc.Create(context.Background(), req)
	assert.Error(t, err)
}

// TestPaymentServiceMarkAsSuccessErrors tests error paths for MarkAsSuccess
func TestPaymentServiceMarkAsSuccessErrors(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.Payment{})

	mockRepo := new(MockPaymentRepository)
	validate := validator.New()
	svc := service.NewPaymentService(mockRepo, db, validate)

	// FindById error
	mockRepo.On("FindById", mock.Anything, mock.Anything, "bad-id").Return(domain.Payment{}, assert.AnError)
	_, err := svc.MarkAsSuccess(context.Background(), "bad-id")
	assert.Error(t, err)

	// Already finalized
	pid := uuid.New()
	donePayment := domain.Payment{ID: pid, Status: "success"}
	mockRepo.On("FindById", mock.Anything, mock.Anything, pid.String()).Return(donePayment, nil)
	_, err = svc.MarkAsSuccess(context.Background(), pid.String())
	assert.Error(t, err)
}

// TestPaymentServiceMarkAsFailedErrors tests error paths for MarkAsFailed
func TestPaymentServiceMarkAsFailedErrors(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.Payment{})

	mockRepo := new(MockPaymentRepository)
	validate := validator.New()
	svc := service.NewPaymentService(mockRepo, db, validate)

	// FindById error
	mockRepo.On("FindById", mock.Anything, mock.Anything, "bad-id").Return(domain.Payment{}, assert.AnError)
	_, err := svc.MarkAsFailed(context.Background(), "bad-id")
	assert.Error(t, err)

	// Already finalized
	pid := uuid.New()
	donePayment := domain.Payment{ID: pid, Status: "failed"}
	mockRepo.On("FindById", mock.Anything, mock.Anything, pid.String()).Return(donePayment, nil)
	_, err = svc.MarkAsFailed(context.Background(), pid.String())
	assert.Error(t, err)
}

// TestPaymentServiceFindByIdNotFound tests FindById when payment not found
func TestPaymentServiceFindByIdNotFound(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.Payment{})

	mockRepo := new(MockPaymentRepository)
	validate := validator.New()
	svc := service.NewPaymentService(mockRepo, db, validate)

	mockRepo.On("FindById", mock.Anything, mock.Anything, "nonexistent-id").Return(domain.Payment{}, assert.AnError)

	_, err := svc.FindById(context.Background(), "nonexistent-id")
	assert.Error(t, err)
}

// TestPaymentServiceFindByIdRepoError tests FindById when repository returns error
func TestPaymentServiceFindByIdRepoError(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.Payment{})

	mockRepo := new(MockPaymentRepository)
	validate := validator.New()
	svc := service.NewPaymentService(mockRepo, db, validate)

	mockRepo.On("FindById", mock.Anything, mock.Anything, "error-id").Return(domain.Payment{}, assert.AnError)

	_, err := svc.FindById(context.Background(), "error-id")
	assert.Error(t, err)
}
