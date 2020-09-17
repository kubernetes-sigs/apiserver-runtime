.PHONY: codegen fix fmt vet lint test tidy

GOBIN := $(shell go env GOPATH)/bin

all: codegen fix fmt vet lint test tidy

fix:
	go fix ./pkg/...
	go fix ./tools/...

fmt:
	go fmt ./pkg/...
	go fmt ./tools/...

lint:
	(which golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint)
	$(GOBIN)/golangci-lint run ./pkg/...
	$(GOBIN)/golangci-lint run ./tools/...

test:
	go test -cover ./...

tidy:
	go mod tidy

vet:
	go vet ./pkg/...
	go vet ./tools/...
