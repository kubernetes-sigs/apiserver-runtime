.PHONY: codegen fix fmt vet lint test sample tidy

GOBIN := $(shell go env GOPATH)/bin

all: codegen fix fmt vet lint test sample tidy

fix:
	go fix ./pkg/...
	go fix ./tools/...

fmt:
	test -z $(go fmt ./pkg/...)
	test -z $(go fmt ./tools/...)

lint:
	(which golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint)
	$(GOBIN)/golangci-lint run ./...

test:
	go test -cover ./...

tidy:
	go mod tidy

vet:
	go vet ./pkg/...
	go vet ./tools/...

sample:
	(cd sample && make)