.PHONY: build run templ css dev clean install-tools

# Install required tools
install-tools:
	go install github.com/a-h/templ/cmd/templ@latest

# Generate templ files
templ:
	go run github.com/a-h/templ/cmd/templ@latest generate

# Build CSS with Tailwind (requires npx/node)
css:
	npx tailwindcss -i ./input.css -o ./static/css/styles.css --minify

# Build the application
build: templ
	go build -o bin/krizzy ./cmd/server

# Run the application
run: templ
	go run ./cmd/server

# Development mode - rebuild and run
dev: templ
	go run ./cmd/server

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f krizzy.db
	find . -name "*_templ.go" -delete

# Watch for changes and rebuild (requires entr or similar)
watch:
	find . -name "*.templ" | entr -r make dev
