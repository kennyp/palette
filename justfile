# Build the CLI
build:
    go build -o bin/palette ./cmd/palette

# Run Tests
test:
    go test -v ./...
