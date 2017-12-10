test:
	go test ./...

%: proofs/%
	go build -o $@ $(shell find $< -maxdepth 1 -name *.go -not -name *_test.go)
