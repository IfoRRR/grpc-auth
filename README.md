# gRPC Auth

Сервис аутентификации на Go с единым JWT-доступом для REST и gRPC API.

**Live demo:** `/grpc-auth/`  
**REST API:** `/api/grpc-auth/`

> Учебный pet-project для портфолио. Не официальный сервис VK.

## Что демонстрирует проект

- выпуск и проверку JWT;
- передачу токена через HTTP Authorization и gRPC metadata;
- unary interceptor для централизованной авторизации gRPC-методов;
- REST и gRPC интерфейсы поверх общей auth-логики;
- PostgreSQL для хранения пользователей;
- Docker Compose и Nginx для запуска демо.

## Архитектура

```text
Client
 ├─ REST + Authorization: Bearer <JWT> → HTTP handlers
 └─ gRPC + metadata               → unary interceptor
                                      ↓
                                Auth service → PostgreSQL
```

## Стек

- Go
- gRPC / Protocol Buffers
- JWT
- PostgreSQL
- Vue
- Docker Compose
- Nginx

## Локальный запуск

```bash
docker compose up -d --build grpc-auth nginx
```

Или запуск Go-сервера напрямую:

```bash
go run ./cmd/server
```

Конфигурация передаётся через переменные окружения. Секреты не должны добавляться в Git.

## Как работает JWT

1. Клиент проходит аутентификацию.
2. Сервис выдаёт access token.
3. REST-клиент передаёт токен в `Authorization: Bearer ...`.
4. gRPC-клиент передаёт токен через metadata.
5. Middleware/interceptor валидирует подпись, срок действия и claims до вызова handler.

## Важно для демо

Для демонстрации используется тестовый пользователь, создаваемый seed-конфигурацией. Пароли и JWT-секреты из production не должны публиковаться в README или коммитах.

## Что можно развить

- refresh token rotation;
- роли и permissions;
- TLS/mTLS для gRPC;
- отзыв токенов;
- интеграционные тесты REST и gRPC;
- observability: metrics, tracing и structured logging.
