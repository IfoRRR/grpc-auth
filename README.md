# grpc-auth

Один JWT для REST и gRPC (unary interceptor + metadata).

Live: `/grpc-auth/` · API: `/api/grpc-auth/`

Demo: `demo` / `demo1234`

## Стек

Go · gRPC · JWT · PostgreSQL · Vue

## Запуск

```bash
docker compose up -d --build grpc-auth nginx
# или
go run ./cmd/server
```
