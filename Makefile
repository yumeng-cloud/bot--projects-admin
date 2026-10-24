.PHONY: dev backend frontend seed test build tidy install

# Start API + Vite together (requires both deps installed)
dev:
	@echo "Starting backend on :8080 and frontend on :5173..."
	@$(MAKE) -j2 run-backend run-frontend

run-backend:
	cd backend && go run ./cmd/server

run-frontend:
	cd frontend && pnpm dev -- --host 0.0.0.0 --port 5173

install:
	cd backend && go mod tidy
	cd frontend && pnpm install

seed:
	cd backend && go run ./cmd/seed

test:
	cd backend && go test ./... -count=1
	cd frontend && pnpm build

build:
	cd backend && go build -o bin/server ./cmd/server
	cd frontend && pnpm build

tidy:
	cd backend && go mod tidy
