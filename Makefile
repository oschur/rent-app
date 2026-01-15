include .env
export 

server-up:
	go run cmd/api/main.go

database-up:
	docker compose up -d && \
	migrate -path internal/migrations -database $(DB_DSN) up 

migrate-down:
	migrate -path internal/migrations -database $(DB_DSN) down

migrate-up:
	migrate -path internal/migrations -database $(DB_DSN) up

