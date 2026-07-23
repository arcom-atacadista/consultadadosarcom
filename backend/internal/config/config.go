// Package config lê a configuração do servidor a partir do ambiente.
package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "3000"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
