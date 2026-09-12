# 01 - Brief Outline

## Problem
Order, stok, payment lintas service harus konsisten. Failure di satu service tidak boleh bikin double-charge atau stok minus. Butuh keandalan ala e-commerce real.

## Scope V1 (MVP Senior)
- API Order idempotent (POST /orders + Idempotency-Key)
- Transactional Outbox pattern (atomic DB + event publish)
- Consumer: Inventory (reserve/release) & Payment (mock success/fail)
- Saga Orchestrator kompensasi jika payment fail -> release stok
- Queue: Kafka (Redpanda) + retry exponential backoff + DLQ
- Redis: distributed lock (order create) + rate limit (token bucket)
- Observability: structured log (slog) + Prometheus metrics + Trace ID

## Non-Goals V1
- Gateway payment beneran (pakai mock)
- Sharding DB, multi-region
- Frontend

## Success Criteria
- Test integration: 100 order concurrent -> tidak duplikat, stok konsisten
- Chaos: kill consumer mid-process -> event reprocess via outbox relay, tidak丢失
- p95 latency < 150ms @ 200 RPS (k6), error < 1%
- Audit log lengkap per state transition

## Actors
Client, Order Service, Inventory Service, Payment Service, Saga Coordinator, Outbox Relay, Kafka
