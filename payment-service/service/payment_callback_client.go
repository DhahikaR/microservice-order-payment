package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"payment-service/models/web"
	"time"

	"github.com/google/uuid"
)

// httpClient is used to perform HTTP requests. It is a variable so tests can substitute it.
var httpClient = &http.Client{}

// SetHTTPClient replaces the internal HTTP client. Use in tests to inject a mock client.
func SetHTTPClient(c *http.Client) {
	if c != nil {
		httpClient = c
		return
	}
	httpClient = &http.Client{}
}

// ResetHTTPClient restores the default HTTP client.
func ResetHTTPClient() {
	httpClient = &http.Client{}
}

func SendPaymentCallbackIntegration(ctx context.Context, callbackURL string, payload web.PaymentCallbackRequest) error {
	return SendPaymentCallback(ctx, callbackURL, payload)
}

// SendPaymentCallback sends a POST request to the callback URL with the payment status payload

func SendPaymentCallback(ctx context.Context, callbackURL string, payload web.PaymentCallbackRequest) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, callbackURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := httpClient
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode >= 400 {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("callback failed with status %d: %s", response.StatusCode, string(body))
	}

	return nil
}

func (service *PaymentServiceImpl) getCallbackURL() string {
	url := os.Getenv("ORDER_CALLBACK_URL")
	return url
}

// getOrderServiceURL returns the order service base URL
func (service *PaymentServiceImpl) getOrderServiceURL() string {
	url := os.Getenv("ORDER_SERVICE_URL")
	return url
}

// fetchOrderAmount fetches order details from order service and validates amount
func (service *PaymentServiceImpl) fetchOrderAmount(ctx context.Context, orderID uuid.UUID) (int64, error) {
	url := fmt.Sprintf("%s/orders/%s", service.getOrderServiceURL(), orderID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return 0, errors.New("order not found")
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("order service returned status %d", resp.StatusCode)
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			TotalAmount int64 `json:"total_amount"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode order response: %w", err)
	}

	return result.Data.TotalAmount, nil
}
