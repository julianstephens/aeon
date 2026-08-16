.PHONY: fmt lint check build debug test

default: check

build:
	@go build -o ./bin/aeon ./cmd/aeon/
	@chmod +x ./bin/aeon

debug:
	@go build -gcflags=all="-N -l" -o ./bin/aeon ./cmd/aeon
	@chmod +x ./bin/aeon

fmt:
	@go mod tidy -v
	@golangci-lint fmt 

lint:
	@go vet ./...
	@golangci-lint run 

test:
	@go test -v -cover -coverpkg ./... -coverprofile=cover.out ./...

check: fmt lint test