package saga

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	platformkafka "go-event-order-service/internal/platform/kafka"
	"go-event-order-service/internal/inventory"
)

type Coordinator struct {
	pool *pgxpool.Pool
	inv  *inventory.Service
	pub  *platformkafka.Publisher
}

func New(pool *pgxpool.Pool, inv *inventory.Service, pub *platformkafka.Publisher) *Coordinator {
	return &Coordinator{pool: pool, inv: inv, pub: pub}
}

func (c *Coordinator) OnInventoryReserved(ctx context.Context, orderID string) {
	log.Printf("saga: inventory reserved %s -> await payment", orderID)
}

func (c *Coordinator) OnPaymentSucceeded(ctx context.Context, orderID string) {
	_, err := c.pool.Exec(ctx, `UPDATE orders SET status='confirmed', updated_at=now() WHERE id=$1`, orderID)
	if err != nil {
		log.Printf("saga confirm fail %s: %v", orderID, err)
		return
	}
	_, _ = c.pool.Exec(ctx, `INSERT INTO order_events(order_id,from_status,to_status,payload) VALUES($1,'pending','confirmed',$2)`, orderID, json.RawMessage(`{"reason":"payment_succeeded"}`))
	_ = c.pub.Publish(ctx, orderID, map[string]any{"event_type": "OrderConfirmed", "order_id": orderID})
	log.Printf("saga: order %s confirmed", orderID)
}

func (c *Coordinator) OnPaymentFailed(ctx context.Context, orderID string) {
	if err := c.inv.Release(ctx, orderID); err != nil {
		log.Printf("saga compensate release fail %s: %v", orderID, err)
	}
	_, _ = c.pool.Exec(ctx, `UPDATE orders SET status='cancelled', updated_at=now() WHERE id=$1`, orderID)
	_, _ = c.pool.Exec(ctx, `INSERT INTO order_events(order_id,from_status,to_status,payload) VALUES($1,'pending','cancelled',$2)`, orderID, json.RawMessage(`{"reason":"payment_failed"}`))
	_ = c.pub.Publish(ctx, orderID, map[string]any{"event_type": "OrderCancelled", "order_id": orderID})
	log.Printf("saga: order %s cancelled + stock released", orderID)
}

func (c *Coordinator) OnInventoryFailed(ctx context.Context, orderID string, reason string) {
	_, _ = c.pool.Exec(ctx, `UPDATE orders SET status='cancelled', updated_at=now() WHERE id=$1`, orderID)
	esc := reason
	if len(esc) > 200 {
		esc = esc[:200]
	}
	payload, _ := json.Marshal(map[string]string{"reason": esc})
	_, _ = c.pool.Exec(ctx, `INSERT INTO order_events(order_id,from_status,to_status,payload) VALUES($1,'pending','cancelled',$2)`, orderID, payload)
	_ = c.pub.Publish(ctx, orderID, map[string]any{"event_type": "OrderCancelled", "order_id": orderID})
	log.Printf("saga: inventory failed %s -> cancelled", orderID)
}
