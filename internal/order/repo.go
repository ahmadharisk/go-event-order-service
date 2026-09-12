package order

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct{ pool *pgxpool.Pool }

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

// CreateWithOutbox: TX orders + order_items + outbox + order_events atomic
func (r *Repo) CreateWithOutbox(ctx context.Context, o Order, items []OrderItem, eventType string, payload any) (Order, error) {
	b, _ := json.Marshal(payload)
	var created Order
	err := r.pool.BeginFunc(ctx, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx,
			`INSERT INTO orders(id,status,idempotency_key,total_cents) VALUES(gen_random_uuid(),'pending',$1,$2) RETURNING id,status,created_at`,
			o.IdempotencyKey, o.TotalCents).Scan(&created.ID, &created.Status, &created.CreatedAt); err != nil {
			return err
		}
		created.IdempotencyKey = o.IdempotencyKey
		created.TotalCents = o.TotalCents
		for _, it := range items {
			if _, err := tx.Exec(ctx,
				`INSERT INTO order_items(order_id,product_id,qty,price_cents) VALUES($1,$2,$3,$4)`,
				created.ID, it.ProductID, it.Qty, it.PriceCents); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO outbox(aggregate_id,event_type,payload) VALUES($1,$2,$3)`,
			created.ID, eventType, b); err != nil {
			return err
		}
		_, err := tx.Exec(ctx,
			`INSERT INTO order_events(order_id,from_status,to_status,payload) VALUES($1,$2,$3,$4)`,
			created.ID, nil, "pending", b)
		return err
	})
	return created, err
}

func (r *Repo) GetByID(ctx context.Context, id string) (Order, error) {
	var o Order
	err := r.pool.QueryRow(ctx,
		`SELECT id,status,idempotency_key,total_cents,created_at FROM orders WHERE id=$1`, id).
		Scan(&o.ID, &o.Status, &o.IdempotencyKey, &o.TotalCents, &o.CreatedAt)
	return o, err
}

func (r *Repo) GetByIdempotencyKey(ctx context.Context, key string) (Order, bool, error) {
	var o Order
	err := r.pool.QueryRow(ctx,
		`SELECT id,status,idempotency_key,total_cents,created_at FROM orders WHERE idempotency_key=$1`, key).
		Scan(&o.ID, &o.Status, &o.IdempotencyKey, &o.TotalCents, &o.CreatedAt)
	if err == pgx.ErrNoRows {
		return Order{}, false, nil
	}
	return o, err == nil, err
}
