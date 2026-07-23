.PHONY: build run dev-backend dev-frontend clean

build:
	cd frontend && bun install && bun run build
	go build -ldflags="-s -w" -o hrsys .

run: build
	./hrsys

dev-backend:
	go run .

dev-frontend:
	cd frontend && bun run dev

clean:
	rm -f hrsys hrsys.db
	rm -rf frontend/dist
