.PHONY: help build test test-race test-coverage lint clean docker-build docker-run docker-push

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  test-race     - Run tests with race detector"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linter"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run the application in Docker"
	@echo "  docker-push   - Push Docker image to registry"

# Build the application
build:
	go build -o bin/gitback ./cmd/gitback

# Run tests
test:
	go test -v ./...

# Run tests with race detector
test-race:
	go test -race -v ./...

# Run tests with coverage report
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run --timeout 5m

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

# Build Docker image
docker-build:
	docker build -t gitback:latest .

# Run the application in Docker
docker-run: docker-build
	docker run --rm -it \
		--env-file .env \
		-v $(pwd)/data:/data \
		gitback:latest

# Push Docker image to registry
docker-push:
	echo "Pushing image to registry..."
	echo "Make sure to update the image name and tag in the Makefile"
	# Example: docker tag gitback:latest your-registry/your-username/gitback:latest
	# docker push your-registry/your-username/gitback:latest
