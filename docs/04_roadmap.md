# 04 - Roadmap (Workflow 9 Langkah)

## Fase 1: Design (Minggu 1) DONE
- [x] 1. Brief (01_brief.md)
- [x] 2. Research (02_research.md)
- [x] 3. Verifikasi sumber
- [x] 4. Design plan (03_design.md) -> revisi sampai ok (posisi sekarang)

## Fase 2: Core Coding (Minggu 2)
- [ ] 5. Code per design: Order TX + Outbox, Chi router, middleware idempotency+rate limit
- urutan: platform/postgres+redis -> order repo/service/handler -> outbox relay -> kafka

## Fase 3: Infra (Minggu 3)
- [ ] Consumer Inventory & Payment + Saga + DLQ + distributed lock + retry backoff

## Fase 4: Productionize (Minggu 4)
- [ ] 6. Tests: testcontainers-go integration (idempotency, outbox, saga, race stock)
- [ ] 7. Run tests -> fix
- [ ] 8. Audit coverage (insert/update/delete, bukan cuma read)
- [ ] 9. Commit
- [ ] Optional: k6 load test 200 RPS, Grafana dashboard, deploy VPS

## Perintah
docker compose -f deploy/docker-compose.yml up -d
make migrate-up
make run
make relay
make worker
make test-integration
make loadtest

## Kriteria Done Senior
- Outbox+Saga+Idempotency+DLQ ada test integration
- golangci-lint 0 issue, go test -race clean
- p95 <150ms, tidak lost update stok
