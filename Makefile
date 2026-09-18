.PHONY: help gen tidy clean up down logs psql migrate-up migrate-down migrate-status migrate-create run-auth

export PATH := $(PATH):$(shell go env GOPATH)/bin

DATABASE_URL ?= postgres://rally:rally@localhost:5433/auth?sslmode=disable
GOOSE_DIR := services/auth/migrations

help:
	@echo "make up                start Postgres"
	@echo "make down              stop Postgres"
	@echo "make logs              tail Postgres logs"
	@echo "make psql              open a Postgres shell (auth database)"
	@echo "make migrate-up        apply pending Auth migrations"
	@echo "make migrate-down      roll back the last Auth migration"
	@echo "make migrate-status    show Auth migration status"
	@echo "make migrate-create name=add_foo   create a new Auth migration file"
	@echo "make run-auth          start the Auth gRPC server"
	@echo "make gen               generate Go from .proto files"
	@echo "make tidy              go mod tidy"
	@echo "make clean             remove generated proto code"

gen:
	mkdir -p gen
	protoc \
	  -I proto \
	  --go_out=gen --go_opt=paths=source_relative \
	  --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
	  auth/v1/auth.proto \
	  group/v1/group.proto \
	  location/v1/location.proto

tidy:
	go mod tidy

clean:
	rm -rf gen/auth gen/group gen/location

up:
	docker compose up -d postgres --wait

down:
	docker compose down

logs:
	docker compose logs -f postgres

psql:
	docker compose exec postgres psql -U rally -d auth

migrate-up: up
	goose -dir $(GOOSE_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir $(GOOSE_DIR) postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir $(GOOSE_DIR) postgres "$(DATABASE_URL)" status

migrate-create:
	mkdir -p $(GOOSE_DIR)
	goose -dir $(GOOSE_DIR) create $(name) sql

run-auth:
	go run ./services/auth/cmd/server
