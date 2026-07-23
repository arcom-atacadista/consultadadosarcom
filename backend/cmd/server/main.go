package main

import (
	"log/slog"
	"net/http"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/config"
	apihttp "github.com/arcom-atacadista/consultadadosarcom/backend/internal/http"
)

func main() {
	cfg := config.Load()
	r := apihttp.NewRouter()

	slog.Info("subindo servidor", "port", cfg.Port)
	if err := http.ListenAndServe("0.0.0.0:"+cfg.Port, r); err != nil {
		slog.Error("servidor encerrado", "erro", err)
	}
}
