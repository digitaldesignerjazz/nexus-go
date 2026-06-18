.PHONY: help build run doctor start-dry start-live clean tidy fmt vet docker-build docker-run

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
	@echo "  make clean          Remove build artifacts"
	@echo "  make docker-build   Build Docker image (nexus-go:latest)"
	@echo "  make docker-run     Run the Docker container interactively"
	@echo ""
	@echo "Examples:"
	@echo "  make doctor"
	@echo "  make start-dry"
	@echo "  make docker-build && make docker-run"
	@echo "  make build && ./bin/nexus-go start --component=mesh"
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

clean:
	rm -rf $(BUILD_DIR)
	@echo "Build artifacts removed."

# Docker targets
 docker-build:
	@echo "Building Docker image nexus-go:latest..."
	docker build -t nexus-go:latest .
	@echo "Docker image built successfully."
	@echo "Run with: make docker-run or docker run --rm -it nexus-go:latest"

 docker-run:
	@echo "Running nexus-go container..."
	docker run --rm -it nexus-go:latest
