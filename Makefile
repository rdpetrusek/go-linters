.PHONY: install lint test
install:
	GO_LINTERS_SOURCE="$(CURDIR)" go run ./cmd/go-linters -dir .bin
lint:
	.bin/golangci-lint run --config .bin/go-linters.yml ./...
test:
	go test ./...
