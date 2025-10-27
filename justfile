# VexDoc MCP Go Server - Justfile
# Usage: just <recipe>

# Default recipe - show available commands
default:
    @just --list

# Install dependencies
deps:
    go mod download
    go mod tidy

# Build the server binary
build:
    go build -o vexdoc-mcp-server ./cmd/server

# Build with optimizations for production
build-prod:
    CGO_ENABLED=0 go build -ldflags="-s -w" -o vexdoc-mcp-server ./cmd/server

# Run the server
run:
    go run ./cmd/server

# Run tests
test:
    go test -v ./...

# Run tests with coverage
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Run tests and show coverage summary
cover:
    go test -cover ./...

# Run tests with coverage
test-coverage:
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Run tests and show coverage in terminal
test-cover:
    go test -v -cover ./...

# Run benchmarks
bench:
    go test -bench=. -benchmem ./...

# Format code
fmt:
    go fmt ./...

# Check if code is formatted (for CI)
fmt-check:
    @test -z "$(gofmt -l .)" || (echo "Code is not formatted. Run 'just fmt'" && gofmt -l . && exit 1)

# Run linter
lint:
    go vet ./...
    @command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping"

# Check for common mistakes
check: fmt lint test

# Clean build artifacts
clean:
    rm -f vexdoc-mcp-server
    rm -f coverage.out coverage.html
    go clean

# Install development tools
install-tools:
    go install honnef.co/go/tools/cmd/staticcheck@latest
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run the server in development mode with verbose output
dev:
    @echo "Starting VexDoc MCP Server in development mode..."
    go run ./cmd/server

# Watch for changes and rebuild (requires entr)
watch:
    @command -v entr >/dev/null 2>&1 || (echo "entr not installed. Install with: apt-get install entr" && exit 1)
    find . -name '*.go' | entr -r just run

# Show project statistics
stats:
    @echo "=== Code Statistics ==="
    @echo "Lines of Go code:"
    @find . -name '*.go' -not -path './vendor/*' | xargs wc -l | tail -1
    @echo "\nGo files:"
    @find . -name '*.go' -not -path './vendor/*' | wc -l
    @echo "\nDependencies:"
    @go list -m all | wc -l

# Docker build
docker-build:
    docker build -t vexdoc-mcp-server:latest .

# Docker run
docker-run:
    docker run --rm -i vexdoc-mcp-server:latest

# Show module dependencies
mod-graph:
    go mod graph

# Update all dependencies
update-deps:
    go get -u ./...
    go mod tidy

# Verify dependencies
verify:
    go mod verify

# Generate documentation
docs:
    @command -v godoc >/dev/null 2>&1 || (echo "godoc not installed. Install with: go install golang.org/x/tools/cmd/godoc@latest" && exit 1)
    @echo "Starting godoc server at http://localhost:6060"
    godoc -http=:6060

# Run security audit
audit:
    @command -v gosec >/dev/null 2>&1 || (echo "gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest" && exit 1)
    gosec ./...

# Full CI pipeline (what runs in CI/CD)
ci: check test-coverage build-prod

# Quick development cycle
quick: fmt test build
