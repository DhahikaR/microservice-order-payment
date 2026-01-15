package test

import (
	"context"
	"errors"
	"testing"

	"order-service/models/domain"
	"order-service/models/web"
	"order-service/service"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockOrderRepository is a testify mock for repository.OrderRepository
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Save(ctx context.Context, tx *gorm.DB, order domain.Order) (domain.Order, error) {
	args := m.Called(ctx, tx, order)
	if args.Get(0) == nil {
		return domain.Order{}, args.Error(1)
	}
	return args.Get(0).(domain.Order), args.Error(1)
}
func (m *MockOrderRepository) Update(ctx context.Context, tx *gorm.DB, order domain.Order) (domain.Order, error) {
	args := m.Called(ctx, tx, order)
	if args.Get(0) == nil {
		return domain.Order{}, args.Error(1)
	}
	return args.Get(0).(domain.Order), args.Error(1)
}
func (m *MockOrderRepository) Delete(ctx context.Context, tx *gorm.DB, orderId string) error {
	args := m.Called(ctx, tx, orderId)
	return args.Error(0)
}
func (m *MockOrderRepository) FindById(ctx context.Context, tx *gorm.DB, orderId string) (domain.Order, error) {
	args := m.Called(ctx, tx, orderId)
	if args.Get(0) == nil {
		return domain.Order{}, args.Error(1)
	}
	return args.Get(0).(domain.Order), args.Error(1)
}
func (m *MockOrderRepository) FindByAll(ctx context.Context, tx *gorm.DB) ([]domain.Order, error) {
	args := m.Called(ctx, tx)
	if args.Get(0) == nil {
		return []domain.Order{}, args.Error(1)
	}
	return args.Get(0).([]domain.Order), args.Error(1)
}

// SUCCESS CONDITION TESTS

// Test Create Endpoint
func TestCreateSuccess(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	req := web.OrderCreateRequest{ItemName: "x", Quantity: 2, Price: 500}
	expected := domain.Order{ID: uuid.New(), ItemName: req.ItemName, Quantity: req.Quantity, Price: req.Price, TotalAmount: req.Price * int64(req.Quantity), Status: "pending"}

	mockRepo.On("Save", mock.Anything, mock.Anything, mock.MatchedBy(func(o domain.Order) bool { return o.ItemName == req.ItemName })).Return(expected, nil)

	got, err := svc.Create(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, expected.ItemName, got.ItemName)
	mockRepo.AssertExpectations(t)
}

// Test Update Endpoint
func TestUpdateSuccess(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	id := uuid.New()

	existing := domain.Order{ID: id, ItemName: "x", Quantity: 1, Price: 100}

	req := web.OrderUpdateRequest{ID: id, ItemName: "y", Quantity: 2, Price: 100}

	mockRepo.On("FindById", mock.Anything, mock.Anything, id.String()).Return(existing, nil)

	updated := existing
	updated.ItemName = req.ItemName
	updated.Quantity = req.Quantity
	updated.Price = req.Price

	mockRepo.On("Update", mock.Anything, mock.Anything, mock.MatchedBy(func(o domain.Order) bool { return o.ItemName == req.ItemName })).Return(updated, nil)

	got, err := svc.Update(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, req.ItemName, got.ItemName)
	assert.Equal(t, req.Quantity, got.Quantity)
	assert.Equal(t, req.Price, got.Price)
	mockRepo.AssertExpectations(t)
}

// Test Delete Endpoint
func TestDeleteSuccess(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	id := uuid.New()
	existing := domain.Order{ID: id, ItemName: "x", Quantity: 1, Price: 100, Status: "pending"}

	mockRepo.On("FindById", mock.Anything, mock.Anything, id.String()).Return(existing, nil)

	mockRepo.On("Delete", mock.Anything, mock.Anything, existing.ID.String()).Return(nil)

	err := svc.Delete(context.Background(), id.String())

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// Test FindById Endpoint
func TestFindByIdSuccess(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	id := uuid.New()
	existing := domain.Order{ID: id, ItemName: "x", Quantity: 1, Price: 100, Status: "pending"}

	mockRepo.On("FindById", mock.Anything, mock.Anything, existing.ID.String()).Return(existing, nil)

	result, err := svc.FindById(context.Background(), existing.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, existing.ID.String(), result.ID.String())

	mockRepo.AssertExpectations(t)
}

// Test FindAll Endpoint
func TestFindAllSuccess(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	existing := []domain.Order{}

	mockRepo.On("FindByAll", mock.Anything, mock.Anything).Return(existing, nil)

	result, err := svc.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, existing, result)

	mockRepo.AssertExpectations(t)
}

// Test ProcessPaymentCallback Endpoint
func TestProcessPaymentCallback_Success(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	id := uuid.New()
	o := domain.Order{ID: id, Status: "pending"}
	mockRepo.On("FindById", mock.Anything, mock.Anything, id.String()).Return(o, nil)
	mockRepo.On("Update", mock.Anything, mock.Anything, mock.MatchedBy(func(ord domain.Order) bool { return ord.Status == "paid" })).Return(domain.Order{ID: id, Status: "paid"}, nil)

	cbReq := web.PaymentCallbackRequest{OrderID: id, PaymentID: uuid.New(), PaymentStatus: "success"}
	got, err := svc.ProcessPaymentCallback(context.Background(), cbReq)
	assert.NoError(t, err)
	assert.Equal(t, "paid", got.Status)
}

// ERROR CONDITION TESTS

// Test Create Endpoint with Validation Error
func TestCreate_ValidationError(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	req := web.OrderCreateRequest{ItemName: "", Quantity: 0, Price: 0}

	assert.Panics(t, func() {
		svc.Create(context.Background(), req)
	})
	mockRepo.AssertNotCalled(t, "Save")
}

// Test Create Endpoint with Repository Error
func TestCreate_RepositoryError(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	req := web.OrderCreateRequest{ItemName: "x", Quantity: 2, Price: 500}

	mockRepo.On("Save", mock.Anything, mock.Anything, mock.AnythingOfType("domain.Order")).Return(domain.Order{}, errors.New("db error"))

	_, err := svc.Create(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())

	mockRepo.AssertExpectations(t)
}

// Test Update Endpoint when Order is Already Paid
func TestUpdate_PaidOrderError(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	id := uuid.New()
	// repository returns already paid order
	mockRepo.On("FindById", mock.Anything, mock.Anything, id.String()).Return(domain.Order{ID: id, Status: "paid"}, nil)

	req := web.OrderUpdateRequest{ID: id, ItemName: "x", Quantity: 1, Price: 100}
	_, err := svc.Update(context.Background(), req)
	assert.Error(t, err)
}

// Test Update Endpoint when Order Invalid Quantity or Price
func TestUpdate_InvalidQuantityOrPrice(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	id := uuid.New()
	mockRepo.On("FindById", mock.Anything, mock.Anything, id.String()).Return(domain.Order{ID: id, Status: "pending"}, nil)

	req := web.OrderUpdateRequest{ID: id, ItemName: "x", Quantity: 0, Price: 100}
	_, err := svc.Update(context.Background(), req)
	assert.Error(t, err)

	req = web.OrderUpdateRequest{ID: id, ItemName: "x", Quantity: 1, Price: 0}
	_, err = svc.Update(context.Background(), req)
	assert.Error(t, err)
}

// Test Delete Endpoint when Order is Already Paid
func TestDelete_PaidOrderError(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	id := uuid.New()
	mockRepo.On("FindById", mock.Anything, mock.Anything, id.String()).Return(domain.Order{ID: id, Status: "paid"}, nil)

	err := svc.Delete(context.Background(), id.String())
	assert.Error(t, err)
}

// Test Delete Endpoint when Order Not Found
func TestDelete_NotFound(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	id := uuid.New()
	mockRepo.On("FindById", mock.Anything, mock.Anything, id.String()).Return(domain.Order{}, errors.New("Not Found"))

	err := svc.Delete(context.Background(), id.String())

	assert.Error(t, err)
	assert.Equal(t, "Not Found", err.Error())

	mockRepo.AssertExpectations(t)
}

// Test Delete Endpoint when Repository Error
func TestDelete_RepositoryError(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	id := uuid.New()
	mockRepo.On("FindById", mock.Anything, mock.Anything, id.String()).Return(domain.Order{}, errors.New("database error"))

	err := svc.Delete(context.Background(), id.String())

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())

	mockRepo.AssertExpectations(t)
}

// Test FindById Endpoint when Order Not Found
func TestFindById_NotFound(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	id := uuid.New().String()

	mockRepo.On("FindById", mock.Anything, mock.Anything, id).Return(domain.Order{}, errors.New("not found"))

	_, err := svc.FindById(context.Background(), id)

	assert.Error(t, err)
	if err == nil {
		t.Fatalf("expected error when finding non-existent order")
	}

	mockRepo.AssertExpectations(t)
}

// Test FindById Endpoint when Repository Error
func TestFindById_RepositoryError(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	id := uuid.New().String()

	mockRepo.On("FindById", mock.Anything, mock.Anything, id).Return(domain.Order{}, errors.New("database error"))

	_, err := svc.FindById(context.Background(), id)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())

	mockRepo.AssertExpectations(t)
}

// Test FindAll Endpoint when Repository Error
func TestFindAll_RepositoryError(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	existing := []domain.Order{}

	mockRepo.On("FindByAll", mock.Anything, mock.Anything).Return(existing, errors.New("database error"))

	_, err := svc.FindAll(context.Background())

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())

	mockRepo.AssertExpectations(t)
}

// Test ProcessPaymentCallback Endpoint when Order Not Found
func TestProcessPaymentCallback_FindOrderError(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	validate := validator.New()
	svc := service.NewOrderService(mockRepo, db, validate)

	id := uuid.New()
	mockRepo.On("FindById", mock.Anything, mock.Anything, id.String()).Return(domain.Order{}, errors.New("not found"))

	cbReq := web.PaymentCallbackRequest{OrderID: id, PaymentID: uuid.New(), PaymentStatus: "success"}
	_, err := svc.ProcessPaymentCallback(context.Background(), cbReq)
	assert.Error(t, err)
}
