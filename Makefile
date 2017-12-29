
GO_FILES := $(shell git ls-files | grep "\.go$$")

PACKAGES := $(shell find $(GO_FILES) | xargs -n1 dirname | sed 's:^:./:'| uniq)

.PHONY: test
test:
	go test ./...

.PHONY: gofmt
gofmt:
	gofmt -e -s -l $(GO_FILES)

.PHONY: govet
govet:
	go vet ./...

.PHONY: golint
golint:
	golint $(PACKAGES)

.PHONY: errcheck
errcheck:
	errcheck -ignoretests $(PACKAGES)

.PHONY: lint
lint: gofmt govet golint errcheck

.PHONY: install
install:
	dep ensure

.PHONY: install-ci
install-ci:
	go get -u github.com/golang/dep/cmd/dep github.com/golang/lint/golint github.com/kisielk/errcheck
	dep ensure

.PHONY: ci
ci: test lint

run/%: proofs/%
	go run $(shell find $< -maxdepth 1 -name *.go -not -name *_test.go)

%: proofs/%
	go build -o $@ $(shell find $< -maxdepth 1 -name *.go -not -name *_test.go)
