package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/auth"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/config"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/db"
	apihttp "github.com/arcom-atacadista/consultadadosarcom/backend/internal/http"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/usuarios"
)

func main() {
	_ = godotenv.Load() // em Docker as variáveis já vêm do compose; isso é só pra dev local
	cfg := config.Load()

	if err := db.Migrate(cfg.DatabaseURL); err != nil {
		slog.Error("falha ao rodar migrations", "erro", err)
		os.Exit(1)
	}
	gdb, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("falha ao conectar no postgres", "erro", err)
		os.Exit(1)
	}

	usuariosRepo := usuarios.NewRepo(gdb)
	authService := auth.NewService(usuariosRepo, cfg.JWTSecret, cfg.SuperAdminEmail)

	r := apihttp.NewRouter(apihttp.Deps{
		JWTSecret:       cfg.JWTSecret,
		AuthHandler:     auth.NewHandler(authService, usuariosRepo),
		UsuariosHandler: usuarios.NewHandler(usuariosRepo),
	})

	slog.Info("subindo servidor", "port", cfg.Port)
	if err := http.ListenAndServe("0.0.0.0:"+cfg.Port, r); err != nil {
		slog.Error("servidor encerrado", "erro", err)
	}
}
