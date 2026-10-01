.PHONY: help up down logs gen migration migrate-up migrate-down psql-auth psql-group redis-cli

export PATH := $(PATH):$(shell go env GOPATH)/bin

COMPOSE := docker compose -f infra/docker-compose.yml

help:
	@echo "make up                                    build and start everything (runs migrations)"
	@echo "make down                                  stop all containers"
	@echo "make logs [service=auth-service]           follow logs (all, or one service)"
	@echo "make gen                                   generate Go from .proto files"
	@echo "make migration service=group name=add_foo  create a new migration file"
	@echo "make migrate-up                            apply pending Auth and Group migrations"
	@echo "make migrate-down                          roll back the last Auth and Group migration"
	@echo "make psql-auth                             open a Postgres shell (auth)"
	@echo "make psql-group                            open a Postgres shell (group)"
	@echo "make redis-cli                             open a Redis shell"

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f $(service)

gen:
	protoc -I proto \
	  --go_out=gen --go_opt=paths=source_relative \
	  --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
	  auth/v1/auth.proto group/v1/group.proto location/v1/location.proto

migration:
	goose -dir services/$(service)/migrations create $(name) sql

migrate-up:
	$(COMPOSE) run --rm --build auth-migrate up
	$(COMPOSE) run --rm --build group-migrate up

migrate-down:
	$(COMPOSE) run --rm --build auth-migrate down
	$(COMPOSE) run --rm --build group-migrate down

psql-auth:
	$(COMPOSE) exec auth-postgres psql -U rally -d auth

psql-group:
	$(COMPOSE) exec group-postgres psql -U rally -d group

redis-cli:
	$(COMPOSE) exec redis redis-cli
