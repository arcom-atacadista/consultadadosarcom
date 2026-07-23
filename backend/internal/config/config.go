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
	ArcomAPIBaseURL string
	ArcomAPIKey     string
}

func Load() Config {
	cfg := Config{
		Port:            getEnv("PORT", "3000"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		RedisURL:        os.Getenv("REDIS_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		SuperAdminEmail: os.Getenv("SUPER_ADMIN_EMAIL"),
		ArcomAPIBaseURL: getEnv("ARCOM_API_BASE_URL", "https://consultacnpj.arcom.com.br"),
		ArcomAPIKey:     os.Getenv("ARCOM_API_KEY"),
	}
	if cfg.JWTSecret == "" {
		slog.Warn("JWT_SECRET não definido no ambiente — login não vai funcionar")
	}
	if cfg.SuperAdminEmail == "" {
		slog.Warn("SUPER_ADMIN_EMAIL não definido — ninguém vira admin automaticamente")
	}
	if cfg.ArcomAPIKey == "" {
		slog.Warn("ARCOM_API_KEY não definido — consulta de CNPJ pela fonte Arcom vai falhar (Brasil API continua funcionando)")
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
