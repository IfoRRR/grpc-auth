package grpcserver

import (
	"context"
	"errors"

	"portfolio/grpc-auth/internal/auth"
	demov1 "portfolio/grpc-auth/gen/demo/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	demov1.UnimplementedAuthServiceServer
	Auth *auth.Service
}

func (s *AuthServer) Register(_ context.Context, req *demov1.RegisterRequest) (*demov1.AuthResponse, error) {
	token, err := s.Auth.Register(req.GetUsername(), req.GetPassword())
	if errors.Is(err, auth.ErrUserExists) {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &demov1.AuthResponse{Token: token, Username: req.GetUsername()}, nil
}

func (s *AuthServer) Login(_ context.Context, req *demov1.LoginRequest) (*demov1.AuthResponse, error) {
	token, err := s.Auth.Login(req.GetUsername(), req.GetPassword())
	if errors.Is(err, auth.ErrInvalidCredentials) {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &demov1.AuthResponse{Token: token, Username: req.GetUsername()}, nil
}
