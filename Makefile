build:
	@go build -o build/watchtower ./cmd/watchtower/main.go

test:
	@go test ./...
