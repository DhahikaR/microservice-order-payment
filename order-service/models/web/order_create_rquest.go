package web

type OrderCreateRequest struct {
	ItemName string `validate:"required"`
	Quantity int    `validate:"required,gt=0"`
	Price    int64  `validate:"required,gt=0"`
}
