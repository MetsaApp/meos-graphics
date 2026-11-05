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
# Local validation. GitHub Actions CI is off, so check runs what
# .github/workflows/test.yml ran on pull requests. scripts/check.sh and the
# pre-push hook (.githooks/pre-push) call it. The tools run pinned through
# go run: templ and swag at their go.mod versions, golangci-lint at the
# README's version (.golangci.yml is a v1 config).
TEMPL_RUN := go run github.com/a-h/templ/cmd/templ@v0.3.865
SWAG_RUN := go run github.com/swaggo/swag/cmd/swag@v1.16.4
GOLANGCI_LINT_RUN := go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2
step = @printf '\n==> %s\n' '$(1)'
# CI ran Go 1.23, and golangci-lint v1.62.2 cannot read a newer standard
# library, so check runs the toolchain go.mod names (go downloads it once).
check: export GOTOOLCHAIN := $(shell sed -n 's/^toolchain //p' go.mod)

.PHONY: check
## check: templ generate, build, test, golangci-lint, and the CLI and Swagger docs are current
check:
	$(call step,tools)
	@command -v go >/dev/null 2>&1 || { echo 'check: go is missing; install Go from https://go.dev/dl/' >&2; exit 1; }
	@command -v grep >/dev/null 2>&1 && echo x | grep -qP x || { echo 'check: scripts/validate-swagger.sh needs GNU grep with -P; sudo apt-get install grep, or brew install grep' >&2; exit 1; }
	$(call step,templ generate)
	$(TEMPL_RUN) generate
	$(call step,go build)
	go build ./...
	$(call step,go test)
	go test ./...
	$(call step,golangci-lint)
	$(GOLANGCI_LINT_RUN) run --timeout=5m --allow-parallel-runners
	$(call step,CLI documentation is current)
	go run ./cmd/generate-docs
	@git diff --exit-code docs/CLI_FLAGS.md || { echo "check: docs/CLI_FLAGS.md is out of date; commit what go run ./cmd/generate-docs wrote" >&2; exit 1; }
	$(call step,Swagger documentation is current)
	@# swag needs a Go file at the root to find the module name.
	echo "package main" > doc.go
	$(SWAG_RUN) init -g cmd/meos-graphics/main.go --parseDependency --parseInternal; status=$$?; rm -f doc.go; exit $$status
	./scripts/validate-swagger.sh
	@git diff --exit-code docs/swagger.json docs/swagger.yaml docs/docs.go || { echo "check: the Swagger docs are out of date; commit what swag init wrote" >&2; exit 1; }
