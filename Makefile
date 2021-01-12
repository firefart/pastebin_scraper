GOPATH := $(or $(GOPATH), $(HOME)/go)

.DEFAULT_GOAL := build

.PHONY: build
build: update test
	go vet ./...
	go fmt ./...
	go build -trimpath .

.PHONY: linux
linux: update test
	GOOS=linux GOARCH=amd64 GO111MODULE=on go build -trimpath .

.PHONY: test
test: update
	go test -v -race ./...

.PHONY: update
update:
	go get -u
	go mod tidy -v

.PHONY: lint
lint:
	@if [ ! -f "$$(go env GOPATH)/bin/golangci-lint" ]; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.34.1; \
	fi
	golangci-lint run ./...
	go mod tidy