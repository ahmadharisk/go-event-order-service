package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go-event-order-service/internal/order"
	"go-event-order-service/internal/platform/config"
	"go-event-order-service/internal/platform/kafka"
	"go-event-order-service/internal/platform/middleware"
	"go-event-order-service/internal/platform/outbox"
	"go-event-order-service/internal/platform/postgres"
	redisclient "go-event-order-service/internal/platform/redis"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	rdb := redisclient.New(cfg.RedisAddr)
	_ = rdb

	// outbox relay async
	if cfg.KafkaBrokers != "" {
		pub := kafka.NewPublisher([]string{cfg.KafkaBrokers}, "order.created")
		relay := outbox.New(pool, pub)
		go relay.Run(ctx)
		defer pub.Close()
	}

	repo := order.NewRepo(pool)
	svc := order.NewService(repo)
	h := order.NewHandler(svc)

	r := chi.NewRouter()
	r.Use(middleware.RateLimit)
	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) { _, _ = w.Write([]byte("ok")) })
	r.Handle("/metrics", promhttp.Handler())
	r.Route("/api/v1", func(r chi.Router) { h.Routes(r) })

	addr := ":" + cfg.HTTPPort
	log.Printf("api listening %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
