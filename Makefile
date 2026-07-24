.PHONY: build dev clean frontend backend install

GOPATH ?= $(shell go env GOPATH)
GOBIN  ?= $(shell go env GOBIN)

install:
	cd frontend && bun install

frontend:
	cd frontend && bun run build

backend:
	go build -o probig-server .

build: frontend backend

dev-frontend:
	cd frontend && bun run dev

dev-backend:
	DEV_MODE=true go run .

dev:
	$(MAKE) -j2 dev-backend dev-frontend

clean:
	rm -f probig-server
	rm -rf frontend/dist
	rm -rf data/

tidy:
	go mod tidy
