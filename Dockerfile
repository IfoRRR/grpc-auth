# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOFLAGS=-buildvcs=false GOOS=linux GOMAXPROCS=1 \
    go build -p 1 -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache wget && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/server /app/server
ENV HTTP_ADDR=:8080 GRPC_ADDR=:9090
EXPOSE 8080 9090
USER app
ENTRYPOINT ["/app/server"]
