.PHONY: build test generate lint clean fmt

build:
	go build -o bin/configforge ./cmd/configforge

test:
	go test ./... -v -race -cover

generate:
	go generate ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ dist/

fmt:
	gofmt -w . && goimports -w .
