package order

import (
	"context"
	"fmt"
)

type Service struct{ repo *Repo }

func NewService(repo *Repo) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, req CreateReq, idemKey string) (Order, bool, error) {
	if idemKey == "" {
		return Order{}, false, fmt.Errorf("Idempotency-Key required")
	}
	if len(req.Items) == 0 {
		return Order{}, false, fmt.Errorf("items empty")
	}
	// idempotency: cek existing
	if existing, found, err := s.repo.GetByIdempotencyKey(ctx, idemKey); err != nil {
		return Order{}, false, err
	} else if found {
		return existing, true, nil
	}
	total := 0
	var items []OrderItem
	for _, it := range req.Items {
		price, ok := PriceMap[it.ProductID]
		if !ok {
			return Order{}, false, fmt.Errorf("unknown product %s", it.ProductID)
		}
		if it.Qty <= 0 {
			return Order{}, false, fmt.Errorf("qty invalid")
		}
		total += price * it.Qty
		items = append(items, OrderItem{ProductID: it.ProductID, Qty: it.Qty, PriceCents: price})
	}
	o := Order{IdempotencyKey: idemKey, TotalCents: total}
	payload := map[string]any{"items": req.Items, "total_cents": total, "idempotency_key": idemKey}
	created, err := s.repo.CreateWithOutbox(ctx, o, items, "OrderCreated", payload)
	return created, false, err
}

func (s *Service) Get(ctx context.Context, id string) (Order, error) {
	return s.repo.GetByID(ctx, id)
}
