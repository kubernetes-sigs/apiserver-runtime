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
	(cd sample && go run ../tools/apiserver-runtime-gen \
    	-g client-gen \
    	-g deepcopy-gen \
    	-g informer-gen \
    	-g lister-gen \
    	-g openapi-gen \
    	--module sigs.k8s.io/apiserver-runtime/sample \
    	--versions sigs.k8s.io/apiserver-runtime/sample/pkg/apis/sample/v1alpha1)

sample-apiserver:
	(cd internal/sample-apiserver && go run ../../tools/apiserver-runtime-gen \
    	-g client-gen \
    	-g deepcopy-gen \
    	-g informer-gen \
    	-g lister-gen \
    	-g openapi-gen \
    	--module sigs.k8s.io/apiserver-runtime/internal/sample-apiserver \
    	--versions sigs.k8s.io/apiserver-runtime/internal/sample-apiserver/pkg/apis/wardle/v1alpha1 \
    	--versions sigs.k8s.io/apiserver-runtime/internal/sample-apiserver/pkg/apis/wardle/v1beta1 )

release-binary:
	mkdir -p bin
	GOOS=linux go build -o bin/apiserver-runtime-gen ./tools/apiserver-runtime-gen
	tar czvf apiserver-runtime-gen-linux.tar.gz bin/apiserver-runtime-gen
	GOOS=darwin go build -o bin/apiserver-runtime-gen ./tools/apiserver-runtime-gen
	tar czvf apiserver-runtime-gen-darwin.tar.gz bin/apiserver-runtime-gen
	GOOS=windows go build -o bin/apiserver-runtime-gen ./tools/apiserver-runtime-gen
	tar czvf apiserver-runtime-gen-windows.tar.gz bin/apiserver-runtime-gen

clean:
	rm -rf bin/
	rm *.tar.gz