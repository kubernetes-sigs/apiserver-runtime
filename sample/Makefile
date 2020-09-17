.PHONY: codegen fix fmt vet lint test tidy

GOBIN := $(shell go env GOPATH)/bin

all: codegen fix fmt vet lint test tidy

docker:
	GOOS=linux GOARCH=amd64 go build -o install/bin/apiserver
	docker build install --tag apiserver-runtime-sample:v0.0.0

install: docker
	kustomize build install | kubectl apply -f -

reinstall: docker
	kustomize build install | kubectl apply -f -
	kubectl delete pods -n sample-system --all

apiserver-logs:
	kubectl logs -l apiserver=true --container apiserver -n sample-system -f --tail 1000

codegen:
	go generate

fix:
	go fix ./...

fmt:
	test -z $(go fmt ./tools/...)

tidy:
	go mod tidy

lint:
	(which golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint)
	$(GOBIN)/golangci-lint run ./...

test:
	go test -cover ./...

vet:
	go vet ./...
