.PHONY: build dev generate css test clean

build: generate css
	@mkdir -p build
	@CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o build/watchtower ./cmd/watchtower

dev:
	@go tool air

generate:
	@go tool templ generate

css:
	@npm run css

test: generate css
	@go test ./...

clean:
	@rm -rf build tmp web/static/app.css
