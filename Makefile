SHELL=/bin/bash

help:
	@egrep -h '^(.+)\:\ ##\ (.+)' ${MAKEFILE_LIST} | column -t -c 2 -s ':#'

run-service:
	@go run -race cmd/rest-server/main.go server

run-all-tests: run-linter run-unit-tests

pre-commit: vendor-deps run-all-tests

restart: stop-compose run-compose-d

redeploy: stop-compose build-image run-compose-d

run-unit-tests:
	@go clean -testcache && go test -v ./... -race

run-pipeline-unit-tests:
	@go clean -testcache && go test -v ./... -race -tags pipeline

run-unit-tests-cover:
	@go test ./... -race -v -coverprofile cover.out && \
	go tool cover -html=cover.out -o cover.html && \
	open file:///$(shell pwd)/cover.html

run-linter:
	@golangci-lint run --deadline=480s --skip-dirs=vendor --tests

# make create-alias alias=some_alias
create-alias:
	@go run -race main.go create $(alias)

install-cli:
	@go install cmd/cli/main.go && mv $(shell go env GOBIN)/{main,payctl}

build-cli-image:
	@docker build -t tigh/payctl:latest -f cli/Dockerfile .

install-linter:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.35.2

install-swagger-gen:
	@go get -d github.com/swaggo/swag/cmd/swag

generate-swagger:
	@swag init --parseVendor --parseDependency --parseInternal -g ./cmd/rest-server/main.go

go-doc-mac:
	@open http://localhost:6060 && \
	godoc -http=:6060

go-doc-linux:
	@xdg-open http://localhost:6060 && \
	godoc -http=:6060

run-compose:
	@docker-compose -f docker-compose.yml -f docker-compose.build.yml  up

run-compose-d:
	@docker-compose -f docker-compose.yml -f docker-compose.build.yml  up -d

run-compose-dev:
	@docker-compose -f docker-compose.yml  -f docker-compose.dev.yml up

run-compose-testnet:
	@docker-compose -f docker-compose.yml  -f docker-compose.testnet.yml up

build-image:
	@docker-compose -f docker-compose.yml -f docker-compose.build.yml build

run-compose-dev-d:
	@docker-compose -f docker-compose.yml -f docker-compose.build.yml -f docker-compose.dev.yml up -d

stop-compose:
	@docker-compose down

vendor-deps:
	@go mod tidy && go mod vendor
