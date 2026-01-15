package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"payment-service/helper"
	"payment-service/models/domain"
	"payment-service/models/web"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestReadFromRequestBodyAndResponseHelpers(t *testing.T) {
	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		var req web.WebResponse
		if err := helper.ReadFromRequestBody(c, &req); err != nil {
			return err
		}
		return helper.ResponseSuccess(c, req.Data)
	})

	// prepare request body
	body := web.WebResponse{Code: 200, Status: "OK", Data: "hello"}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestPaymentResponseHelpers(t *testing.T) {
	payment := domain.Payment{ID: uuid.New(), OrderID: uuid.New(), Amount: 1000, Status: "paid"}

	resp := helper.ToPaymentResponse(payment)

	assert.Equal(t, resp.ID, payment.ID)
	assert.Equal(t, resp.OrderID, payment.OrderID)
	assert.Equal(t, resp.Amount, payment.Amount)
	assert.Equal(t, resp.Status, payment.Status)
}

func TestResponseHelpersStatusCodes(t *testing.T) {
	app := fiber.New()
	app.Get("/bad", func(c *fiber.Ctx) error { return helper.BadRequest(c, "bad") })
	app.Get("/notfound", func(c *fiber.Ctx) error { return helper.NotFound(c, "not") })
	app.Get("/internal", func(c *fiber.Ctx) error { return helper.InternalServerError(c, "err") })

	r, _ := app.Test(httptest.NewRequest("GET", "/bad", nil))
	assert.Equal(t, 400, r.StatusCode)
	r, _ = app.Test(httptest.NewRequest("GET", "/notfound", nil))
	assert.Equal(t, 404, r.StatusCode)
	r, _ = app.Test(httptest.NewRequest("GET", "/internal", nil))
	assert.Equal(t, 500, r.StatusCode)
}

func TestResponseAndReadFromRequestBody(t *testing.T) {
	app := fiber.New()

	app.Post("/parse", func(c *fiber.Ctx) error {
		var req web.PaymentCreateRequest
		if err := helper.ReadFromRequestBody(c, &req); err != nil {
			return helper.BadRequest(c, err.Error())
		}
		return helper.ResponseSuccess(c, req)
	})

	// valid JSON
	body, _ := json.Marshal(web.PaymentCreateRequest{OrderID: uuid.New(), Amount: 1000})
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

func TestCommitOrRollback(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&domain.Payment{})

	// commit case
	tx := db.Begin()
	tx.Create(&domain.Payment{
		ID:      uuid.New(),
		OrderID: uuid.New(),
		Amount:  1000,
		Status:  "paid"})
	helper.CommitOrRollback(tx)
	var count int64
	db.Model(&domain.Payment{}).Count(&count)
	assert.EqualValues(t, int64(1), count)

	// rollback case: simulate panic inside function
	func() {
		tx2 := db.Begin()
		defer func() {
			if err := recover(); err != nil {
			}
		}()
		defer helper.CommitOrRollback(tx2)
		tx2.Create(&domain.Payment{
			ID:      uuid.New(),
			OrderID: uuid.New(),
			Amount:  1000,
			Status:  "paid"})
		panic("error")
	}()

	db.Model(&domain.Payment{}).Count(&count)
	assert.EqualValues(t, int64(1), count)
}
