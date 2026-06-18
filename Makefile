.PHONY: help build run doctor start-dry start-live clean tidy fmt vet lint docker-build docker-push docker-run

# Nexus-go Makefile
# Go-based orchestrator for the full Nexus ecosystem (mesh + blockchain + AI swarms + prototypes)
# Part of Esslinger & Co. technology stack

BINARY_NAME := nexus-go
BUILD_DIR := bin

help:
	@echo ""
	@echo "Nexus-go — Go-based Nexus Ecosystem Orchestrator"
	@echo "=================================================="
	@echo ""
	@echo "Available targets:"
	@echo "  make help           Show this help message"
	@echo "  make build          Build optimized binary into ./bin/"
	@echo "  make run            Run via 'go run' (quick development)"
	@echo "  make doctor         Run environment & Nexus readiness checks"
	@echo "  make start-dry      Safe dry-run preview of full stack startup (recommended first step)"
	@echo "  make start-live     LIVE execution mode — makes real system changes (use with extreme caution)"
	@echo "  make tidy           Run go mod tidy"
	@echo "  make fmt            Format Go code"
	@echo "  make vet            Run go vet"
	@echo "  make lint           Run golangci-lint"
	@echo "  make clean          Remove build artifacts"
	@echo "  make docker-build   Build Docker image locally"
	@echo "  make docker-push    Build and push multi-arch image to GHCR (requires login)"
	@echo "  make docker-run     Run the Docker container interactively"
	@echo ""
	@echo "Examples:"
	@echo "  make doctor"
	@echo "  make start-dry"
	@echo "  make lint"
	@echo "  make docker-build"
	@echo "  make docker-push"
	@echo ""

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Binary created: $(BUILD_DIR)/$(BINARY_NAME)"

run:
	go run main.go $(filter-out $@,$(MAKECMDGOALS))

 doctor:
	go run main.go doctor

start-dry:
	go run main.go start --component=all

start-live:
	@echo ""
	@echo "WARNING: LIVE EXECUTION MODE"
	@echo "This will attempt to make real changes to your system (Yggdrasil, Docker, configs, etc.)."
	@echo "Current implementation is still mostly placeholder — no destructive actions yet."
	@echo ""
	@echo "Press Ctrl+C to abort, or press Enter to continue..."
	@read -r dummy
	go run main.go start --all --execute --force

 tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)
	@echo "Build artifacts removed."

# Docker targets
 docker-build:
	@echo "Building Docker image nexus-go:latest..."
	docker build -t nexus-go:latest .
	@echo "Docker image built successfully."

 docker-push:
	@echo "Building and pushing multi-arch Docker image to GHCR..."
	docker buildx build --platform linux/amd64,linux/arm64 \
		-t ghcr.io/digitaldesignerjazz/nexus-go:latest \
		-t ghcr.io/digitaldesignerjazz/nexus-go:$(git describe --tags --abbrev=0) \
		--push .
	@echo "Docker image pushed to GHCR."

 docker-run:
	@echo "Running nexus-go container..."
	docker run --rm -it nexus-go:latest
