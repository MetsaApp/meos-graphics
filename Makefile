.PHONY: build run test clean templ css css-watch swagger docs

# Build the application
build: templ css
	go build -o bin/meos-graphics ./cmd/meos-graphics

# Run the application
run: templ css
	go run ./cmd/meos-graphics

# Run in simulation mode
run-sim: templ css
	go run ./cmd/meos-graphics --simulation

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f web/static/css/styles.css
	rm -f internal/web/templates/*_templ.go

# Generate Go code from templ files
templ:
	templ generate

# Build CSS with Tailwind
css:
	npm run build-css

# Watch CSS changes
css-watch:
	npm run watch-css

# Generate Swagger documentation
swagger:
	swag init -g cmd/meos-graphics/main.go --parseDependency --parseInternal

# Validate Swagger documentation
swagger-validate: swagger
	./scripts/validate-swagger.sh

# Update all documentation
docs: swagger

# Install dependencies
deps:
	go mod download
	npm install
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# Development mode - run with hot reload
dev:
	@echo "Starting development mode..."
	@echo "Run 'make css-watch' in another terminal for CSS hot reload"
	templ generate --watch &
	air

# Full build - templ, CSS, swagger, and Go binary
full-build: templ css swagger build