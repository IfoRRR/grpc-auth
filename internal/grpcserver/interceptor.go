package grpcserver

import (
	"context"
	"strings"

	"portfolio/grpc-auth/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthUnaryInterceptor проверяет JWT в metadata "authorization".
func AuthUnaryInterceptor(authSvc *auth.Service) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if info.FullMethod == "/demo.v1.AuthService/Register" || info.FullMethod == "/demo.v1.AuthService/Login" {
			return handler(ctx, req)
		}
		user, err := userFromContext(ctx, authSvc)
		if err != nil {
			return nil, err
		}
		return handler(ContextWithUser(ctx, user), req)
	}
}

func userFromContext(ctx context.Context, authSvc *auth.Service) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}
	user, err := authSvc.Validate(strings.Join(vals, " "))
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid token")
	}
	return user, nil
}

type ctxKey string

const userKey ctxKey = "user"

func ContextWithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userKey).(string)
	return v, ok && v != ""
}
