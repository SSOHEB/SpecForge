.PHONY: build test test-golden test-integration bench release-build docker-build docker-run generate lint clean fmt

build:
	go build -o bin/configforge ./cmd/configforge

test:
	go test ./internal/... ./cmd/configforge/cmd -v -race -cover

test-golden:
	go test ./tests/golden/... -v

test-integration:
	go test ./tests/integration/... -v

bench:
	go test ./tests/integration/... -bench=. -benchmem -run=^$

release-build:
	go run scripts/build_release.go

docker-build:
	docker build -t configforge:latest .

docker-run:
	docker run --rm configforge:latest defaults -m examples/http-server/metadata.yaml

generate:
	go generate ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ dist/

fmt:
	gofmt -w . && goimports -w .
