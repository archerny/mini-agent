.PHONY: all build build-web build-all test run clean dev

# Build Go backend only (no embedded frontend)
build:
	go build -o bin/mini-agent ./cmd/server

# Build frontend
build-web:
	cd web && npm run build

# Build single binary with embedded frontend
build-all: build-web
	go build -tags embedfrontend -o bin/mini-agent ./cmd/server

# Run Go tests
test:
	go test ./... -v -count=1

# Run the server (backend only)
run: build
	./bin/mini-agent

# Run single binary with embedded frontend
run-all: build-all
	./bin/mini-agent

# Run frontend dev server
dev-web:
	cd web && npm run dev

# Run both backend and frontend (development)
dev:
	@echo "Starting backend and frontend..."
	@trap 'kill 0' EXIT; \
		$(MAKE) run & \
		$(MAKE) dev-web & \
		wait

# TypeScript type check
typecheck:
	cd web && npx tsc --noEmit

# Install frontend dependencies
install-web:
	cd web && npm install

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf web/dist/
	rm -rf web/node_modules/
