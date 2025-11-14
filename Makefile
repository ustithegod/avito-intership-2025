start:
	docker compose up --build

generate-jwt:
	go run ./cmd/token_generator

