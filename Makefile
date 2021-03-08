.DEFAULT_GOAL := build

.PHONY: build
build:
	go vet ./...
	go fmt ./...
	go build -trimpath .

.PHONY: linux
linux: update test
	GOOS=linux GOARCH=amd64 go build -trimpath .

.PHONY: test
test: update
	go test -v -race ./...

.PHONY: update
update:
	go get -u
	go mod tidy -v

.PHONY: lint
lint:
	"$$(go env GOPATH)/bin/golangci-lint" run ./...
	go mod tidy

.PHONY: lint-update
lint-update:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	$$(go env GOPATH)/bin/golangci-lint --version

.PHONY: lint-docker
lint-docker:
	docker pull golangci/golangci-lint:latest
	docker run --rm -v $$(pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run
