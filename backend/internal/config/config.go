// Package config lê a configuração do servidor a partir do ambiente.
package config

import (
	"log/slog"
	"os"
)

type Config struct {
	Port            string
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	SuperAdminEmail string
}

func Load() Config {
	cfg := Config{
		Port:            getEnv("PORT", "3000"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		RedisURL:        os.Getenv("REDIS_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		SuperAdminEmail: os.Getenv("SUPER_ADMIN_EMAIL"),
	}
	if cfg.JWTSecret == "" {
		slog.Warn("JWT_SECRET não definido no ambiente — login não vai funcionar")
	}
	if cfg.SuperAdminEmail == "" {
		slog.Warn("SUPER_ADMIN_EMAIL não definido — ninguém vira admin automaticamente")
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
