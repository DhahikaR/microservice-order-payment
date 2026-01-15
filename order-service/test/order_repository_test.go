package test

import (
	"context"
	"testing"

	"order-service/models/domain"
	"order-service/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestOrderRepositoryCRUD(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// AutoMigrate would generate Postgres-specific SQL (gen_random_uuid()) from tags.
	// Create a simple sqlite-compatible table instead to run tests.
	createSQL := `CREATE TABLE orders (
        id TEXT PRIMARY KEY,
        item_name TEXT,
        quantity INTEGER,
        price INTEGER,
        total_amount INTEGER,
        status TEXT,
        payment_id TEXT,
        created_at DATETIME,
        updated_at DATETIME,
        deleted_at DATETIME
    );`
	err = db.Exec(createSQL).Error
	assert.NoError(t, err)

	repo := repository.NewOrderRepository(db)

	tx := db.Begin()

	id := uuid.New()
	o := domain.Order{ID: id, ItemName: "item", Quantity: 2, Price: 1000, TotalAmount: 2000, Status: "pending"}

	saved, err := repo.Save(context.Background(), tx, o)
	assert.NoError(t, err)
	assert.Equal(t, o.ID, saved.ID)

	found, err := repo.FindById(context.Background(), tx, id.String())
	assert.NoError(t, err)
	assert.Equal(t, id, found.ID)

	// update
	found.ItemName = "new"
	updated, err := repo.Update(context.Background(), tx, found)
	assert.NoError(t, err)
	assert.Equal(t, "new", updated.ItemName)

	// find all
	all, err := repo.FindByAll(context.Background(), tx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(all), 1)

	// delete (soft delete)
	err = repo.Delete(context.Background(), tx, id.String())
	assert.NoError(t, err)

	// after delete, find should return error
	_, err = repo.FindById(context.Background(), tx, id.String())
	assert.Error(t, err)

	tx.Commit()
}
