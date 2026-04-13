run:
	go run cmd/api/main.go

build:
	go build -o bin/rmp-api cmd/api/main.go

migrate:
	go run cmd/migrate/main.go

test:
	go test ./...

docker-up:
	docker compose up -d

docker-down:
	docker compose down