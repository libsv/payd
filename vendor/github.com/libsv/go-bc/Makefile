SHELL=/bin/bash

help:
	@egrep -h '^(.+)\:\ ##\ (.+)' ${MAKEFILE_LIST} | column -t -c 2 -s ':#'

run-all-tests: run-linter run-unit-tests

pre-commit: vendor-deps run-all-tests

run-unit-tests:
	@go clean -testcache && go test -v ./... -race

run-unit-tests-cover:
	@go test ./... -race -v -coverprofile cover.out && \
	go tool cover -html=cover.out -o cover.html && \
	open file:///$(shell pwd)/cover.html

run-linter:
	@golangci-lint run --deadline=480s --skip-dirs=vendor --tests

install-linter:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.35.2

go-doc-mac:
	@open http://localhost:6060 && \
	godoc -http=:6060

go-doc-linux:
	@xdg-open http://localhost:6060 && \
	godoc -http=:6060

vendor-deps:
	@go mod tidy && go mod vendor

tag: ## Generate a new tag and push (tag version=0.0.0)
	@run-all-tests
	@git tag -a v$(version) -m "Pending full release..."
	@git push origin v$(version)
	@git fetch --tags -f
