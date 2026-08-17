.PHONY: fmt lint check build debug test pack

default: check

pack:
	@rm -rf out && mkdir -p out
	@go run ./cmd/mapviz/main.go render --layer elevation --output out/elevation.png
	@go run ./cmd/mapviz/main.go render --layer moisture --output out/moisture.png
	@go run ./cmd/mapviz/main.go render --layer fertility --output out/fertility.png
	@go run ./cmd/mapviz/main.go render --layer terrain --output out/terrain.png


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