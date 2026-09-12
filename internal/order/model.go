package order

import "time"

type ItemReq struct {
	ProductID string `json:"product_id"`
	Qty       int    `json:"qty"`
}

type CreateReq struct {
	Items []ItemReq `json:"items"`
}

type Order struct {
	ID             string    `json:"id"`
	Status         string    `json:"status"`
	IdempotencyKey string    `json:"idempotency_key"`
	TotalCents     int       `json:"total_cents"`
	CreatedAt      time.Time `json:"created_at"`
}

type OrderItem struct {
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Qty       int    `json:"qty"`
	PriceCents int   `json:"price_cents"`
}

// price mock, real ambil dari inventory/catalog
var PriceMap = map[string]int{
	"PROD-001": 15000,
	"PROD-002": 25000,
	"PROD-003": 5000,
}
