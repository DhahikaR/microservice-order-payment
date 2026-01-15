package test

import (
	"context"
	"testing"
	"time"

	"payment-service/models/domain"
	"payment-service/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPaymentRepositoryCRUD(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// migrate schema
	err = db.AutoMigrate(&domain.Payment{})
	assert.NoError(t, err)

	repo := repository.NewPaymentRepository(db)

	tx := db.Begin()

	// Create
	pid := uuid.New()
	oid := uuid.New()
	p := domain.Payment{ID: pid, OrderID: oid, Amount: 1000, Provider: "x", Status: "pending"}

	saved, err := repo.Save(context.Background(), tx, p)
	assert.NoError(t, err)
	assert.Equal(t, p.ID, saved.ID)

	found, err := repo.FindById(context.Background(), tx, pid.String())
	assert.NoError(t, err)
	assert.Equal(t, pid, found.ID)

	// update status
	now := time.Now()
	found.Status = "success"
	found.PaidAt = &now
	updated, err := repo.UpdateStatus(context.Background(), tx, found)
	assert.NoError(t, err)
	assert.Equal(t, "success", updated.Status)

	// FindOrderById should find by order id
	byOrder, err := repo.FindOrderById(context.Background(), tx, oid.String())
	assert.NoError(t, err)
	assert.Equal(t, oid, byOrder.OrderID)

	tx.Commit()
}

func TestFindById_NotFound(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	_ = db.AutoMigrate(&domain.Payment{})

	repo := repository.NewPaymentRepository(db)
	tx := db.Begin()
	defer tx.Rollback()

	// search for non-existing id
	_, err = repo.FindById(context.Background(), tx, "non-existent-id")
	assert.Error(t, err)
}

func TestUpdateStatus_NotFound(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	_ = db.AutoMigrate(&domain.Payment{})

	repo := repository.NewPaymentRepository(db)
	tx := db.Begin()
	defer tx.Rollback()

	// update non-existing payment - GORM may not return error for zero rows updated
	p := domain.Payment{ID: uuid.New(), Status: "success"}
	_, err = repo.UpdateStatus(context.Background(), tx, p)
	// UpdateStatus should not error even if nothing updated; verify record still not found
	assert.NoError(t, err)
	_, err = repo.FindById(context.Background(), tx, p.ID.String())
	assert.Error(t, err)
}
