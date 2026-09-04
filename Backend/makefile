include .envrc

MIGRATIONS_PATH = ./migrations

.PHONY: migration
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migration-up
migration-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_MIGRATOR_ADDR) up

.PHONY: migration-down
migration-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_MIGRATOR_ADDR) down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: gen-docs
gen-docs:
	@swag init -g cmd/src/main.go -d cmd,internal && swag fmt

# Catch-all: evita que Make falle cuando se pasan argumentos extra
# a migration/migration-down (ej. `make migration create_users`)
%:
	@: