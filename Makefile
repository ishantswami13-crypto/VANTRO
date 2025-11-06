run:
	go run ./cmd/api

tidy:
	go mod tidy

migrate:
	psql "$$DATABASE_URL" -f migrations/001_init.sql

build:
	go build -o bin/vantro ./cmd/api

docker-build:
	docker build -t vantro:dev .

docker-run:
	docker run --env-file .env -p 8080:8080 vantro:dev
