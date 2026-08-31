.PHONY: fmt fmt-check vet test check demo

fmt:
	gofmt -w ./cmd ./internal

fmt-check:
	@test -z "$$(gofmt -l ./cmd ./internal)" || (echo "Run 'make fmt'" && gofmt -l ./cmd ./internal && exit 1)

vet:
	go vet ./...

test:
	go test -race ./...

check: fmt-check vet test

demo:
	go run ./cmd/elasticchaind demo-scaling
