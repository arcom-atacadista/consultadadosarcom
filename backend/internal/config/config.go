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
	GeoapifyAPIKey  string
	Trace360BaseURL string
	Trace360APIKey  string
	GroqAPIKey      string
	GroqAPIURL      string
	TavilyAPIKey    string
	TavilyAPIURL    string
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
		GeoapifyAPIKey:  os.Getenv("GEOAPIFY_API_KEY"),
		Trace360BaseURL: getEnv("TRACE360_BASE_URL", "https://trace360ai.arcom.com.br/api/v1"),
		Trace360APIKey:  os.Getenv("TRACE360_API_KEY"),
		GroqAPIKey:      os.Getenv("GROQ_API_KEY"),
		GroqAPIURL:      os.Getenv("GROQ_API_URL"),
		TavilyAPIKey:    os.Getenv("TAVILY_API_KEY"),
		TavilyAPIURL:    os.Getenv("TAVILY_API_URL"),
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
	if cfg.GeoapifyAPIKey == "" {
		slog.Warn("GEOAPIFY_API_KEY não definido — geocodificação (fachada/Street View) vai falhar")
	}
	if cfg.Trace360APIKey == "" {
		slog.Warn("TRACE360_API_KEY não definido — enriquecimento vai falhar")
	}
	if cfg.GroqAPIKey == "" {
		slog.Warn("GROQ_API_KEY não definido — insight/ranking/chat de IA vão falhar")
	}
	if cfg.TavilyAPIKey == "" {
		slog.Warn("TAVILY_API_KEY não definido — busca na web (insight/chat) vai falhar")
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
