package test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"payment-service/models/web"
	"payment-service/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// errTransport is a RoundTripper that always returns the provided error.
type errTransport struct{ err error }

func (e errTransport) RoundTrip(req *http.Request) (*http.Response, error) { return nil, e.err }

// TestSendPaymentCallback ensures the callback client posts payload
func TestSendPaymentCallbackIntegration(t *testing.T) {
	received := make(chan web.PaymentCallbackRequest, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify method and headers
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
			return
		}

		var p web.PaymentCallbackRequest
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		received <- p
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Do not rely on side-effect env var in this test; pass URL directly.
	payload := web.PaymentCallbackRequest{PaymentID: uuid.New(), OrderID: uuid.New(), PaymentStatus: "success"}
	err := service.SendPaymentCallback(context.Background(), srv.URL, payload)
	assert.NoError(t, err)

	select {
	case got := <-received:
		assert.Equal(t, payload.PaymentID, got.PaymentID)
		assert.Equal(t, payload.OrderID, got.OrderID)
		assert.Equal(t, payload.PaymentStatus, got.PaymentStatus)
	case <-time.After(1 * time.Second):
		t.Fatal("callback not received within timeout")
	}
}

// TestSendPaymentCallbackInvalidURL tests handling of invalid URL
func TestSendPaymentCallbackInvalidURL(t *testing.T) {
	invalidURL := "http://invalid-url"
	payload := web.PaymentCallbackRequest{PaymentID: uuid.New(), OrderID: uuid.New(), PaymentStatus: "success"}
	err := service.SendPaymentCallback(context.Background(), invalidURL, payload)
	assert.Error(t, err)
}

// TestSendPaymentCallbackServerError tests handling of server error response
func TestSendPaymentCallbackServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	payload := web.PaymentCallbackRequest{PaymentID: uuid.New(), OrderID: uuid.New(), PaymentStatus: "success"}
	err := service.SendPaymentCallback(context.Background(), srv.URL, payload)
	assert.Error(t, err)
}

// TestSendPaymentCallbackClientError simulates an http.Client that fails to perform requests.
func TestSendPaymentCallbackClientError(t *testing.T) {
	service.SetHTTPClient(&http.Client{Transport: errTransport{err: errors.New("net error")}})
	defer service.ResetHTTPClient()

	payload := web.PaymentCallbackRequest{PaymentID: uuid.New(), OrderID: uuid.New(), PaymentStatus: "success"}
	err := service.SendPaymentCallback(context.Background(), "http://example.com", payload)
	assert.Error(t, err)
}
