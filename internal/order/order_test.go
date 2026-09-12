package order_test

import (
	"testing"
)

// flags: go test -tags=integration butuh testcontainers-go
// doc: test nyata butuh Postgres + Redis + Redpanda via testcontainers
// kasus wajib: idempotency duplikat, outbox publish after crash, saga kompensasi, race stock 100 concurrent

func TestIdempotencyDuplicate(t *testing.T) {
	t.Skip("need testcontainers: Postgres+Redis; implement via testcontainers-go pgx")
	// 1. POST /orders key=abc -> 201
	// 2. POST ulang key=abc -> 200 cached, count orders ==1
}

func TestOutboxRelayPublishes(t *testing.T) {
	t.Skip("need Kafka Redpanda container: insert outbox pending -> relay tick -> consume order.created")
}

func TestSagaCompensateOnPaymentFail(t *testing.T) {
	t.Skip("force payment mock failed -> assert stock release + orders status cancelled + order_events row")
}

func TestConcurrentStockNoLostUpdate(t *testing.T) {
	t.Skip("100 goroutine reserve same product stock=50 -> hanya 50 sukses, sisa inventory.failed, tidak stock minus")
}
