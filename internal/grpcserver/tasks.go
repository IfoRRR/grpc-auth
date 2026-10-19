package grpcserver

import (
	"context"

	"portfolio/grpc-auth/internal/store"
	demov1 "portfolio/grpc-auth/gen/demo/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskServer struct {
	demov1.UnimplementedTaskServiceServer
	Store *store.Store
}

func (s *TaskServer) CreateTask(ctx context.Context, req *demov1.CreateTaskRequest) (*demov1.CreateTaskResponse, error) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title required")
	}
	t := s.Store.Create(user, req.GetTitle())
	return &demov1.CreateTaskResponse{Task: toProto(t)}, nil
}

func (s *TaskServer) GetTask(ctx context.Context, req *demov1.GetTaskRequest) (*demov1.GetTaskResponse, error) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}
	t, err := s.Store.Get(user, req.GetId())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &demov1.GetTaskResponse{Task: toProto(t)}, nil
}

func (s *TaskServer) ListTasks(ctx context.Context, _ *demov1.ListTasksRequest) (*demov1.ListTasksResponse, error) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no user in context")
	}
	list := s.Store.List(user)
	out := make([]*demov1.Task, 0, len(list))
	for _, t := range list {
		out = append(out, toProto(t))
	}
	return &demov1.ListTasksResponse{Tasks: out}, nil
}

func toProto(t store.Task) *demov1.Task {
	return &demov1.Task{Id: t.ID, Title: t.Title, Done: t.Done}
}
