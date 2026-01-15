package helper

import (
	"payment-service/models/domain"
	"payment-service/models/web"
)

func ToPaymentResponse(payment domain.Payment) web.PaymentResponse {
	return web.PaymentResponse{
		ID:      payment.ID,
		OrderID: payment.OrderID,
		Amount:  payment.Amount,
		Status:  payment.Status,
		PaidAt:  payment.PaidAt,
	}
}
