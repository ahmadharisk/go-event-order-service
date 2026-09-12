package inventory

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct{ pool *pgxpool.Pool }

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

func (r *Repo) Reserve(ctx context.Context, orderID string) error {
	return r.pool.BeginFunc(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `SELECT product_id, qty FROM order_items WHERE order_id=$1`, orderID)
		if err != nil {
			return err
		}
		defer rows.Close()
		type item struct{ pid string; qty int }
		var items []item
		for rows.Next() {
			var it item
			if err := rows.Scan(&it.pid, &it.qty); err != nil {
				return err
			}
			items = append(items, it)
		}
		for _, it := range items {
			tag, err := tx.Exec(ctx,
				`UPDATE inventory SET stock=stock-$1, version=version+1, updated_at=now() WHERE product_id=$2 AND stock>=$1`,
				it.qty, it.pid)
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return fmt.Errorf("insufficient stock %s", it.pid)
			}
		}
		return nil
	})
}

func (r *Repo) Release(ctx context.Context, orderID string) error {
	return r.pool.BeginFunc(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `SELECT product_id, qty FROM order_items WHERE order_id=$1`, orderID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var pid string; var qty int
			if err := rows.Scan(&pid, &qty); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `UPDATE inventory SET stock=stock+$1, version=version+1 WHERE product_id=$2`, qty, pid); err != nil {
				return err
			}
		}
		return nil
	})
}
