.PHONY: up down down-volumes build

up:
	docker compose up -d

down:
	docker compose down

down-volumes:
	docker compose down -v

build:
	go build -o bin/bot ./cmd/bot/