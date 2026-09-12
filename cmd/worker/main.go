package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"go-event-order-service/internal/inventory"
	"go-event-order-service/internal/payment"
	"go-event-order-service/internal/platform/config"
	platformkafka "go-event-order-service/internal/platform/kafka"
	"go-event-order-service/internal/platform/postgres"
	"go-event-order-service/internal/saga"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	brokers := strings.Split(cfg.KafkaBrokers, ",")
	invSvc := inventory.NewService(inventory.NewRepo(pool))
	paySvc := payment.NewService(payment.NewRepo(pool))
	confirmPub := platformkafka.NewPublisher(brokers, "order.events")
	defer confirmPub.Close()
	coord := saga.New(pool, invSvc, confirmPub)

	go consumeOrderCreated(ctx, brokers, invSvc, coord)
	go consumeInventoryEvents(ctx, brokers, paySvc)
	go consumePaymentEvents(ctx, brokers, coord)

	log.Println("worker running: order.created -> inventory -> payment -> saga")
	select {}
}

func consumeOrderCreated(ctx context.Context, brokers []string, inv *inventory.Service, coord *saga.Coordinator) {
	r := platformkafka.NewReader(brokers, "order.created", "inventory-group")
	defer r.Close()
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("order.created read: %v", err)
			time.Sleep(time.Second)
			continue
		}
		orderID := extractOrderID(m.Value)
		if orderID == "" {
			log.Printf("order.created: missing order_id")
			continue
		}
		if err := inv.Reserve(ctx, orderID); err != nil {
			log.Printf("inventory reserve fail %s: %v", orderID, err)
			pub := platformkafka.NewPublisher(brokers, "inventory.failed")
			_ = pub.Publish(ctx, orderID, map[string]any{"order_id": orderID, "reason": err.Error()})
			pub.Close()
			coord.OnInventoryFailed(ctx, orderID, err.Error())
			continue
		}
		pub := platformkafka.NewPublisher(brokers, "inventory.reserved")
		_ = pub.Publish(ctx, orderID, map[string]any{"order_id": orderID})
		pub.Close()
		coord.OnInventoryReserved(ctx, orderID)
	}
}

func consumeInventoryEvents(ctx context.Context, brokers []string, pay *payment.Service) {
	r := platformkafka.NewReader(brokers, "inventory.reserved", "payment-group")
	defer r.Close()
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		orderID := extractOrderID(m.Value)
		status, perr := pay.Process(ctx, orderID)
		if perr != nil {
			log.Printf("payment err %s: %v", orderID, perr)
		}
		topic := "payment.succeeded"
		if status == "failed" {
			topic = "payment.failed"
		}
		pub := platformkafka.NewPublisher(brokers, topic)
		_ = pub.Publish(ctx, orderID, map[string]any{"order_id": orderID, "status": status})
		pub.Close()
	}
}

func consumePaymentEvents(ctx context.Context, brokers []string, coord *saga.Coordinator) {
	go func() {
		r := kafka.NewReader(kafka.ReaderConfig{Brokers: brokers, Topic: "payment.succeeded", GroupID: "saga-group"})
		defer r.Close()
		for {
			m, err := r.ReadMessage(ctx)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			coord.OnPaymentSucceeded(ctx, extractOrderID(m.Value))
		}
	}()
	r := kafka.NewReader(kafka.ReaderConfig{Brokers: brokers, Topic: "payment.failed", GroupID: "saga-group"})
	defer r.Close()
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		coord.OnPaymentFailed(ctx, extractOrderID(m.Value))
	}
}

func extractOrderID(b []byte) string {
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return ""
	}
	if v, ok := m["order_id"].(string); ok {
		return v
	}
	if p, ok := m["payload"].(map[string]any); ok {
		if v, ok := p["order_id"].(string); ok {
			return v
		}
	}
	if v, ok := m["aggregate_id"].(string); ok {
		return v
	}
	return ""
}
