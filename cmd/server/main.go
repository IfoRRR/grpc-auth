// Command server — REST (:8080) + gRPC (:9090) с JWT-авторизацией.
// Браузер ходит в REST; /probe доказывает, что gRPC принимает тот же Bearer в metadata.
package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"

	"portfolio/grpc-auth/internal/auth"
	"portfolio/grpc-auth/internal/config"
	"portfolio/grpc-auth/internal/grpcserver"
	"portfolio/grpc-auth/internal/rest"
	"portfolio/grpc-auth/internal/store"
	demov1 "portfolio/grpc-auth/gen/demo/v1"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
	if len(cfg.JWTSecret) < 16 {
		log.Fatal("JWT_SECRET не задан или короче 16 символов; задай GRPC_AUTH_JWT_SECRET в .env")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(5)
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("postgres:", err)
	}

	authSvc, err := auth.New(cfg.JWTSecret, sqlDB)
	if err != nil {
		log.Fatal(err)
	}
	_, _ = authSvc.Register("demo", "demo1234") // ignore exists

	st, err := store.Open(sqlDB)
	if err != nil {
		log.Fatal(err)
	}

	restSrv := &rest.Server{Auth: authSvc, Store: st, GRPCAddr: cfg.GRPCAddr}

	go func() {
		log.Println("REST", cfg.HTTPAddr)
		log.Fatal(http.ListenAndServe(cfg.HTTPAddr, restSrv.Routes()))
	}()

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal(err)
	}
	gs := grpc.NewServer(grpc.UnaryInterceptor(grpcserver.AuthUnaryInterceptor(authSvc)))
	demov1.RegisterAuthServiceServer(gs, &grpcserver.AuthServer{Auth: authSvc})
	demov1.RegisterTaskServiceServer(gs, &grpcserver.TaskServer{Store: st})
	log.Println("gRPC", cfg.GRPCAddr)
	log.Fatal(gs.Serve(lis))
}
