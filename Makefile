.PHONY: fmt fmt-check vet test check build demo localnet-init localnet-start localnet-stop localnet-smoke

fmt:
	gofmt -w ./cmd ./internal

fmt-check:
	@test -z "$$(gofmt -l ./cmd ./internal)" || (echo "Run 'make fmt'" && gofmt -l ./cmd ./internal && exit 1)

vet:
	go vet ./...

test:
	go test -race ./...

check: fmt-check vet test

build:
	mkdir -p build
	go build -o build/elasticchaind ./cmd/elasticchaind

demo:
	go run ./cmd/elasticchaind demo-scaling

localnet-init: build
	bash scripts/localnet-init.sh

localnet-start:
	bash scripts/localnet-start.sh

localnet-stop:
	bash scripts/localnet-stop.sh

localnet-smoke: build
	bash scripts/localnet-smoke.sh
