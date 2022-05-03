SHELL=/bin/bash

help:
	@egrep -h '^(.+)\:\ ##\ (.+)' ${MAKEFILE_LIST} | column -t -c 2 -s ':#'

run-service:
	@go run -race cmd/bip270-server/main.go server

run-all-tests: run-linter run-unit-tests

pre-commit: vendor-deps run-all-tests

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

install-linter:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)bin v1.44.2

install-swagger-gen:
	@go get -d github.com/swaggo/swag/cmd/swag

generate-swagger:
	@swag init --parseVendor --parseDependency --parseInternal -g ./cmd/server/main.go

go-doc-mac:
	@open http://localhost:6060 && \
	godoc -http=:6060

go-doc-linux:
	@xdg-open http://localhost:6060 && \
	godoc -http=:6060

create-pc-account:
	@docker exec peerchannels ./SPVChannels.API.Rest -createaccount admin username password 2>/dev/null || true

run-compose: create-pc-account
	@docker-compose -f docker-compose.yml -f docker-compose.build.yml  up

run-compose-d: create-pc-account
	@docker-compose -f docker-compose.yml -f docker-compose.build.yml  up -d

run-compose-dev: create-pc-account
	@docker-compose -f docker-compose.yml  -f docker-compose.dev.yml up

run-compose-local: create-pc-account
	@docker-compose -f docker-compose.yml  -f docker-compose.local.yml up

run-compose-faucet: create-pc-account
	@docker-compose -f docker-compose.yml  -f docker-compose.faucet.yml up

run-compose-local-d: create-pc-account
	@docker-compose -f docker-compose.yml  -f docker-compose.local.yml up -d

run-compose-testnet: create-pc-account
	@docker-compose -f docker-compose.yml  -f docker-compose.testnet.yml up

build-image:
	@docker-compose -f docker-compose.yml -f docker-compose.build.yml build

run-compose-dev-d: create-pc-account
	@docker-compose -f docker-compose.yml -f docker-compose.build.yml -f docker-compose.dev.yml up -d

stop-compose:
	@docker-compose down

vendor-deps:
	@go mod tidy && go mod vendor
