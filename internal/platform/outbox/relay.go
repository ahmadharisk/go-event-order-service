package outbox

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	platformkafka "go-event-order-service/internal/platform/kafka"
)

type Relay struct {
	pool      *pgxpool.Pool
	publisher *platformkafka.Publisher
	interval  time.Duration
}

func New(pool *pgxpool.Pool, pub *platformkafka.Publisher) *Relay {
	return &Relay{pool: pool, publisher: pub, interval: 2 * time.Second}
}

func (r *Relay) Run(ctx context.Context) {
	tick := time.NewTicker(r.interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if err := r.tick(ctx); err != nil {
				log.Printf("outbox tick: %v", err)
			}
		}
	}
}

func (r *Relay) tick(ctx context.Context) error {
	rows, err := r.pool.Query(ctx,
		`SELECT id, aggregate_id, event_type, payload FROM outbox WHERE status='pending' ORDER BY created_at LIMIT 20 FOR UPDATE SKIP LOCKED`)
	if err != nil {
		return err
	}
	defer rows.Close()
	type msg struct {
		ID          string
		AggregateID string
		EventType   string
		Payload     json.RawMessage
	}
	var msgs []msg
	for rows.Next() {
		var m msg
		if err := rows.Scan(&m.ID, &m.AggregateID, &m.EventType, &m.Payload); err != nil {
			return err
		}
		msgs = append(msgs, m)
	}
	for _, m := range msgs {
		var payload any
		_ = json.Unmarshal(m.Payload, &payload)
		if err := r.publisher.Publish(ctx, m.AggregateID, map[string]any{"event_type": m.EventType, "payload": payload}); err != nil {
			continue
		}
		_, _ = r.pool.Exec(ctx, `UPDATE outbox SET status='published', published_at=now() WHERE id=$1`, m.ID)
	}
	return nil
}
