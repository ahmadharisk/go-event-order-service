.PHONY: migrate-up migrate-down run relay worker test-integration lint loadtest

migrate-up:
	migrate -path migrations -database "${DATABASE_URL}" up

migrate-down:
	migrate -path migrations -database "${DATABASE_URL}" down 1

run:
	go run ./cmd/api

relay:
	go run ./cmd/relay

worker:
	go run ./cmd/worker

test-integration:
	go test ./... -race -count=1 -tags=integration

lint:
	golangci-lint run ./...

loadtest:
	k6 run deploy/k6/script.js
