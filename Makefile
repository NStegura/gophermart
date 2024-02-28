$(eval CNT_API := $(shell docker ps -f name=gophermart-api -q | wc -l | awk '{print $1}'))

.PHONY: up
up:
ifeq ($(CNT_API),0)
	docker-compose up --build --no-recreate --detach; sleep 5
endif

.PHONY: bash
bash:
	make up
	docker-compose exec gophermart-api /bin/sh

.PHONY: down
down:
	docker-compose down --remove-orphans --rmi local

.PHONY: buildall
buildall: buildgophermart buildaccrual

.PHONY: buildgophermart
buildapi: ## Build api app
	go build -o ./cmd/gophermart/gophermart cmd/gophermart/main.go

.PHONY: rundb
rundb:
	docker run --name gophermart -e POSTGRES_USER=usr -e POSTGRES_PASSWORD=psswrd -e POSTGRES_DB=metrics -p 54323:5432 -d postgres:14.2

.PHONY: migrate
migrate:
	goose -dir=internal/repo/internal/db/migrations postgres "host=localhost port=54323 user=usr password=psswrd dbname=gophermart sslmode=disable" up

.PHONY: rollbackmigrations
rollbackmigrations:
	goose -dir=internal/repo/internal/db/migrations postgres "host=localhost port=54323 user=usr password=psswrd dbname=gophermart sslmode=disable" reset

.PHONY: swagger
swagger: ## generate swagger files
	swag fmt
	swag init --parseDependency -g internal/app/gophermartapi/server.go

## TESTS

MOCKS_DESTINATION=mocks
.PHONY: mocks
# put the files with interfaces you'd like to mock in prerequisites
# wildcards are allowed
mocks: ./internal/app/gophermartapi/iauth.go \
       ./internal/app/gophermartapi/ibusiness.go \
       ./internal/services/business/irepository.go \
       ./internal/services/jobs/accrualsync/irepository.go \
       ./internal/services/jobs/accrualsync/iaccrualcli.go
	@echo "Generating mocks..."
	@rm -rf $(MOCKS_DESTINATION)
	@for file in $^; do mockgen -source=$$file -destination=$(MOCKS_DESTINATION)/$$file; done

.PHONY: test
test:
	go install gotest.tools/gotestsum@latest
	gotestsum --format pkgname -- -coverprofile=cover.out ./...

## LINTERS
GOLANGCI_LINT_CACHE?=/tmp/praktikum-golangci-lint-cache

.PHONY: fmt
fmt:
	go fmt ./...
	goimports -w -local github.com/NStegura/gophermart ./cmd
	goimports -w -local github.com/NStegura/gophermart ./internal

.PHONY: lint
lint:
	golangci-lint run -c .golangci.yml --out-format=colored-line-number --sort-results

.PHONY: lint-report
lint-report: _golangci-lint-rm-unformatted-report

.PHONY: _golangci-lint-reports-mkdir
_golangci-lint-reports-mkdir:
	mkdir -p ./golangci-lint

.PHONY: _golangci-lint-run
_golangci-lint-run: _golangci-lint-reports-mkdir
	-docker run --rm \
    -v $(shell pwd):/app \
    -v $(GOLANGCI_LINT_CACHE):/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.55.2 \
        golangci-lint run \
            -c .golangci.yml \
	> ./golangci-lint/report-unformatted.json

.PHONY: _golangci-lint-format-report
_golangci-lint-format-report: _golangci-lint-run
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json

.PHONY: _golangci-lint-rm-unformatted-report
_golangci-lint-rm-unformatted-report: _golangci-lint-format-report
	rm ./golangci-lint/report-unformatted.json

.PHONY: lint-clean
golangci-lint-clean:
	sudo rm -rf ./golangci-lint
