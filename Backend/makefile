include .envrc

MIGRATIONS_PATH = ./migrations

.PHONY: migration
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migration-up
migration-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) up

.PHONY: migration-down
migration-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: gen-docs
gen-docs:
	@swag init -g cmd/src/main.go -d cmd,internal && swag fmt

# Flags que usan `make test` y `make test-cover`. Se pueden sobreescribir:
#   make test TEST_FLAGS="-run TestUserService -v"
TEST_FLAGS ?= -race -cover

.PHONY: test
test:
	@go vet ./...
	@go test $(TEST_FLAGS) ./...

.PHONY: test-cover
test-cover:
	@go test -race -covermode=atomic -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

.PHONY: clean-cover
clean-cover:
	@rm -f coverage.out

# Catch-all: evita que Make falle cuando se pasan argumentos extra
# a migration/migration-down (ej. `make migration create_users`)
%:
	@: