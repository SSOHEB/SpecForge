FROM mcr.microsoft.com/devcontainers/go:2.1-bookworm

# Install golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.0

# Install goimports (used by 'make fmt')
RUN go install golang.org/x/tools/cmd/goimports@latest
