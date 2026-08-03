# Stage 1: Build
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X configforge/cmd/configforge/cmd.version=$(git describe --tags --always || echo dev) -X configforge/cmd/configforge/cmd.commit=$(git rev-parse --short HEAD || echo none) -X configforge/cmd/configforge/cmd.date=$(date -Iseconds || echo unknown)" \
    -o /configforge ./cmd/configforge

# Stage 2: Final minimal image
FROM alpine:3.20
COPY --from=builder /configforge /configforge
ENTRYPOINT ["/configforge"]
