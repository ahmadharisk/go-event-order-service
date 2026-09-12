# Go Event-Driven Order & Payment Service
### Portofolio Level Senior Backend (Go)

**Tujuan:** Latih skill senior: concurrency, consistency (Outbox+Saga), idempotency, queue reliability, observability.

**Stack:** Go + Chi + Postgres 16 + Redis 7 + Kafka (Redpanda) + Docker + OTEL + Prometheus/Grafana + k6

**Fase:** Lihat docs/01_brief.md -> 02_research.md -> 03_design.md -> 04_roadmap.md

**Quick start:**
```bash
docker compose -f deploy/docker-compose.yml up -d
make migrate-up && make run
make test-integration
```
