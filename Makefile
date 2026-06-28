.PHONY: all frontend backend check docker-build docker-run backend-test

all: frontend backend

frontend:
	cd apps/web && npm install && npm run build

check:
	cd apps/web && npm run check

backend:
	cd server && go build -o bin/server ./cmd/server

backend-test:
	cd server && go test ./...

docker-build:
	docker build -t koalanotes:dev .

docker-run:
	docker compose up -d

dev-frontend:
	cd apps/web && npm run dev

dev-backend:
	cd server && go run ./cmd/server
