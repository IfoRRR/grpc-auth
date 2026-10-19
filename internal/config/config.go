package config

import "os"

type Config struct {
	HTTPAddr    string
	GRPCAddr    string
	JWTSecret   string
	GRPCProbe   bool
}

func Load() Config {
	return Config{
		HTTPAddr:  env("HTTP_ADDR", ":8080"),
		GRPCAddr:  env("GRPC_ADDR", ":9090"),
		JWTSecret: env("JWT_SECRET", ""),
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
