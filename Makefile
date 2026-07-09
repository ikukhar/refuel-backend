.PHONY: run build test lint migrate-up migrate-down

run:
	$(shell go env GOPATH)/bin/air

build:
	go build -o refuel-backend ./cmd/server

test:
	go test ./...

lint:
	golangci-lint run

migrate-up:
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

migrate-down:
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
