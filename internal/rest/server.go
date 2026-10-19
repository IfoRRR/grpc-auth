package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	demov1 "portfolio/grpc-auth/gen/demo/v1"
	"portfolio/grpc-auth/internal/auth"
	"portfolio/grpc-auth/internal/store"
)

type Server struct {
	Auth     *auth.Service
	Store    *store.Store
	GRPCAddr string
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("POST /auth/register", s.register)
	mux.HandleFunc("POST /auth/login", s.login)
	mux.Handle("POST /tasks", s.requireAuth(http.HandlerFunc(s.createTask)))
	mux.Handle("GET /tasks", s.requireAuth(http.HandlerFunc(s.listTasks)))
	mux.Handle("GET /tasks/{id}", s.requireAuth(http.HandlerFunc(s.getTask)))
	mux.Handle("POST /probe", s.requireAuth(http.HandlerFunc(s.probeGRPC)))
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"bad json"}`, http.StatusBadRequest)
		return
	}
	token, err := s.Auth.Register(body.Username, body.Password)
	if errors.Is(err, auth.ErrUserExists) {
		http.Error(w, `{"error":"username taken"}`, http.StatusConflict)
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"token": token, "username": body.Username})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"bad json"}`, http.StatusBadRequest)
		return
	}
	token, err := s.Auth.Login(body.Username, body.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token, "username": body.Username})
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	user := userFromRequest(r)
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		http.Error(w, `{"error":"title required"}`, http.StatusBadRequest)
		return
	}
	t := s.Store.Create(user, body.Title)
	writeJSON(w, http.StatusCreated, t)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	user := userFromRequest(r)
	t, err := s.Store.Get(user, r.PathValue("id"))
	if errors.Is(err, store.ErrNotFound) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	user := userFromRequest(r)
	writeJSON(w, http.StatusOK, map[string]any{"tasks": s.Store.List(user)})
}

// probeGRPC — для live demo: тот же JWT уходит в gRPC metadata, ListTasks через stub.
func (s *Server) probeGRPC(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r)
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(s.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		http.Error(w, `{"error":"grpc dial failed"}`, http.StatusBadGateway)
		return
	}
	defer conn.Close()

	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx = metadata.NewOutgoingContext(ctx, md)

	client := demov1.NewTaskServiceClient(conn)
	resp, err := client.ListTasks(ctx, &demov1.ListTasksRequest{})
	if err != nil {
		http.Error(w, `{"error":"grpc call failed"}`, http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"via":   "grpc",
		"count": len(resp.GetTasks()),
		"tasks": resp.GetTasks(),
	})
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		user, err := s.Auth.Validate(token)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(contextWithUser(r.Context(), user)))
	})
}

type restUserKey struct{}

func contextWithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, restUserKey{}, user)
}

func userFromRequest(r *http.Request) string {
	if v, ok := r.Context().Value(restUserKey{}).(string); ok {
		return v
	}
	return ""
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	return strings.TrimPrefix(h, "Bearer ")
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
