.PHONY: help gen tidy clean up down logs psql psql-group \
	migrate-up migrate-down migrate-status migrate-create \
	migrate-group-up migrate-group-down migrate-group-status migrate-group-create \
	run-auth run-group

export PATH := $(PATH):$(shell go env GOPATH)/bin

AUTH_DATABASE_URL ?= postgres://rally:rally@localhost:5433/auth?sslmode=disable
GROUP_DATABASE_URL ?= postgres://rally:rally@localhost:5434/group?sslmode=disable
DATABASE_URL ?= $(AUTH_DATABASE_URL)
AUTH_GOOSE_DIR := services/auth/migrations
GROUP_GOOSE_DIR := services/group/migrations
GOOSE_DIR := $(AUTH_GOOSE_DIR)

help:
	@echo "make up                      start Auth and Group Postgres"
	@echo "make down                    stop Postgres"
	@echo "make logs                    tail Postgres logs"
	@echo "make psql                    open a Postgres shell (auth)"
	@echo "make psql-group              open a Postgres shell (group)"
	@echo "make migrate-up              apply pending Auth migrations"
	@echo "make migrate-down            roll back the last Auth migration"
	@echo "make migrate-status          show Auth migration status"
	@echo "make migrate-create name=add_foo   create a new Auth migration file"
	@echo "make migrate-group-up        apply pending Group migrations"
	@echo "make migrate-group-down      roll back the last Group migration"
	@echo "make migrate-group-status    show Group migration status"
	@echo "make migrate-group-create name=add_foo   create a new Group migration file"
	@echo "make run-auth                start the Auth gRPC server"
	@echo "make run-group               start the Group gRPC server"
	@echo "make gen                     generate Go from .proto files"
	@echo "make tidy                    go mod tidy"
	@echo "make clean                   remove generated proto code"

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
	docker compose up -d auth-postgres group-postgres --wait

down:
	docker compose down

logs:
	docker compose logs -f auth-postgres group-postgres

psql:
	docker compose exec auth-postgres psql -U rally -d auth

psql-group:
	docker compose exec group-postgres psql -U rally -d group

migrate-up: up
	goose -dir $(AUTH_GOOSE_DIR) postgres "$(AUTH_DATABASE_URL)" up

migrate-down:
	goose -dir $(AUTH_GOOSE_DIR) postgres "$(AUTH_DATABASE_URL)" down

migrate-status:
	goose -dir $(AUTH_GOOSE_DIR) postgres "$(AUTH_DATABASE_URL)" status

migrate-create:
	mkdir -p $(AUTH_GOOSE_DIR)
	goose -dir $(AUTH_GOOSE_DIR) create $(name) sql

migrate-group-up: up
	goose -dir $(GROUP_GOOSE_DIR) postgres "$(GROUP_DATABASE_URL)" up

migrate-group-down:
	goose -dir $(GROUP_GOOSE_DIR) postgres "$(GROUP_DATABASE_URL)" down

migrate-group-status:
	goose -dir $(GROUP_GOOSE_DIR) postgres "$(GROUP_DATABASE_URL)" status

migrate-group-create:
	mkdir -p $(GROUP_GOOSE_DIR)
	goose -dir $(GROUP_GOOSE_DIR) create $(name) sql

run-auth:
	go run ./services/auth/cmd/server

run-group:
	go run ./services/group/cmd/server
