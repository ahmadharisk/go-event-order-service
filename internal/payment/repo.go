package payment

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct{ pool *pgxpool.Pool }

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

func (r *Repo) Insert(ctx context.Context, orderID string, amount int, status string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO payments(order_id, amount_cents, status) VALUES($1,$2,$3) ON CONFLICT(order_id) DO UPDATE SET status=EXCLUDED.status`, orderID, amount, status)
	return err
}

func (r *Repo) FindOrderAmount(ctx context.Context, orderID string) (int, error) {
	var amt int
	err := r.pool.QueryRow(ctx, `SELECT total_cents FROM orders WHERE id=$1`, orderID).Scan(&amt)
	return amt, err
}
