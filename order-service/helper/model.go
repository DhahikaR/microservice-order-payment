package helper

import (
	"order-service/models/domain"
	"order-service/models/web"
)

func ToOrderResponse(order domain.Order) web.OrderResponse {
	return web.OrderResponse{
		Id:          order.ID,
		ItemName:    order.ItemName,
		Quantity:    order.Quantity,
		Price:       order.Price,
		TotalAmount: order.TotalAmount,
		Status:      order.Status,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}
}

func ToOrderResponses(orders []domain.Order) []web.OrderResponse {
	var orderResponses []web.OrderResponse
	for _, order := range orders {
		orderResponses = append(orderResponses, ToOrderResponse(order))
	}

	return orderResponses
}
