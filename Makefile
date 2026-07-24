.PHONY: build build-frontend build-backend dev dev-backend dev-frontend clean run

BINARY_NAME=probig
FRONTEND_DIR=frontend
EMBED_DIR=internal/frontend/dist

build: build-frontend build-backend
	@echo "Build complete: $(BINARY_NAME)"

build-frontend:
	@echo "Building frontend..."
	cd $(FRONTEND_DIR) && bun install && bun run build
	rm -rf $(EMBED_DIR)
	cp -r $(FRONTEND_DIR)/dist $(EMBED_DIR)

build-backend:
	@echo "Building backend..."
	go mod tidy
	go build -o $(BINARY_NAME) ./cmd/server/

dev-backend:
	@echo "Starting backend dev server..."
	go run ./cmd/server/

dev-frontend:
	@echo "Starting frontend dev server..."
	cd $(FRONTEND_DIR) && bun run dev

dev:
	@echo "Starting all dev servers..."

clean:
	rm -f $(BINARY_NAME)
	rm -rf $(EMBED_DIR)
	rm -f hr.db

run:
	./$(BINARY_NAME)
