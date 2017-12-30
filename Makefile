GO_FILES := $(shell git ls-files | grep "\.go$$")

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	gometalinter --vendor \
		--exclude bindata.go \
		--disable-all \
		--enable gofmt \
		--enable golint \
		--enable vet \
		--enable errcheck \
		./...

.PHONY: ci
ci: test lint

run/%: proofs/%
	go run $(shell find $< -maxdepth 1 -name *.go -not -name *_test.go)

%: proofs/%
	go build -o $@ $(shell find $< -maxdepth 1 -name *.go -not -name *_test.go)
