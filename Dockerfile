# Stage 1: Build
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X github.com/SSOHEB/codrao/cmd/codrao/cmd.version=$(git describe --tags --always || echo dev) -X github.com/SSOHEB/codrao/cmd/codrao/cmd.commit=$(git rev-parse --short HEAD || echo none) -X github.com/SSOHEB/codrao/cmd/codrao/cmd.date=$(date -Iseconds || echo unknown)" \
    -o /codrao ./cmd/codrao

# Stage 2: Final minimal image
FROM alpine:3.20
COPY --from=builder /codrao /codrao
ENTRYPOINT ["/codrao"]
