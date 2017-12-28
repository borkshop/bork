test:
	go test ./...

run/%: proofs/%
	go run $(shell find $< -maxdepth 1 -name *.go -not -name *_test.go)

%: proofs/%
	go build -o $@ $(shell find $< -maxdepth 1 -name *.go -not -name *_test.go)
