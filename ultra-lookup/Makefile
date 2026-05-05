GO ?= go
.PHONY: run server load-http observability down test bench tidy

run:
	$(GO) run ./cmd/ultra-lookup \
		-customers 6000000 \
		-accounts 12000000 \
		-shards 256 \
		-lookups 5000000 \
		-workers 16 \
		-hit-ratio 95 \
		-sample-every 2048 \
		-seed 42 \
		-rh-load 0.90

bench:
	$(GO) test ./internal/lookup -bench BenchmarkLookup -benchmem -run ^$

server:
	$(GO) run ./cmd/server \
		-customers 6000000 \
		-accounts 12000000 \
		-shards 256 \
		-rh-load 0.90

load-http:
	$(GO) run ./cmd/loadgen \
		-url http://localhost:8080 \
		-duration 60s \
		-workers 16 \
		-customers 6000000 \
		-accounts 12000000 \
		-hit-ratio 95 \
		-append-every 1000 \
		-sample-every 1024

observability:
	docker compose up --build

down:
	docker compose down

test:
	$(GO) test ./...

tidy:
	$(GO) mod tidy
