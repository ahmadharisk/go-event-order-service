# 02 - Research Best Practices & Sumber

## 1. Transactional Outbox
- Tulis orders + outbox dalam 1 TX Postgres, relay terpisah publish ke Kafka. Cegah lost event.
- Sumber: microservices.io/patterns/communication/outbox, Chris Richardson - Microservices Patterns
- Go: pgx + sqlc, polling SELECT ... FOR UPDATE SKIP LOCKED atau LISTEN/NOTIFY

## 2. Saga Pattern (Orchestration)
- Pilih Orchestration untuk V1: traceable, mudah kompensasi. Choreography chain panjang susah debug.
- Sumber: temporal.io/blog/saga, Microsoft Learn - Saga pattern

## 3. Idempotency
- Header Idempotency-Key: uuid + DB UNIQUE + cache response 24j
- Sumber: stripe.com/docs/api/idempotent_requests, IETF draft-ietf-httpapi-idempotency-key-header

## 4. Kafka Reliability
- acks=all, enable.idempotence=true, manual commit setelah sukses, retry topic + DLQ
- Sumber: kafka.apache.org/documentation, Confluent - Designing Event-Driven Systems
- Dev: Redpanda (Kafka-compatible, 1 container, tanpa ZK)

## 5. Postgres Concurrency
- UPDATE stock pakai version optimistic atau SELECT ... FOR UPDATE. Hindari lost update.
- READ COMMITTED + row lock cukup; SERIALIZABLE hanya untuk ledger kritis
- Sumber: postgresql.org/docs/current/transaction-iso.html

## 6. Redis Lock & Rate Limit
- Lock: SET NX PX + Lua unlock (go-redsync). TTL 5-10s.
- Rate limit: Token Bucket via INCR+EXPIRE atau golang.org/x/time/rate + Redis global
- Sumber: redis.io/docs/manual/patterns/distlock/

## 7. Observability Go
- slog JSON + TraceID via context + X-Request-ID
- Metrics: prometheus/client_golang
- Tracing: go.opentelemetry.io/otel
- Sumber: opentelemetry.io/docs/languages/go

## 8. Testing Senior
- Prefunctional/integration: testcontainers-go untuk Postgres/Redis/Redpanda real
- Sumber: testcontainers.com/modules/postgres

## Checklist Verifikasi (workflow langkah 3)
- [ ] Cek link masih live
- [ ] Cocokkan versi: pgx v5, chi v5, redpanda latest
- [ ] Validasi status outbox: pending -> published -> failed
