.PHONY: build generate css test clean

build: generate css
	@mkdir -p build
	@go build -o build/watchtower ./cmd/watchtower

generate:
	@go tool templ generate

css:
	@npm run css

test: generate
	@go test ./...

clean:
	@rm -rf build public/app.css
