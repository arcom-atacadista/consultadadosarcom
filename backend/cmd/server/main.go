package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/auth"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/cnpj"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/config"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/db"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/geo"
	apihttp "github.com/arcom-atacadista/consultadadosarcom/backend/internal/http"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/prospeccao"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/redis"
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
	rdb, err := redis.Connect(cfg.RedisURL)
	if err != nil {
		slog.Error("falha ao conectar no redis", "erro", err)
		os.Exit(1)
	}

	usuariosRepo := usuarios.NewRepo(gdb)
	authService := auth.NewService(usuariosRepo, cfg.JWTSecret, cfg.SuperAdminEmail)

	// Um único ArcomClient compartilhado — cnpj e prospeccao falam com a
	// mesma API (mesmo token, mesma reautenticação em 401).
	arcomClient := cnpj.NewArcomClient(cfg.ArcomAPIBaseURL, cfg.ArcomAPIKey)

	cnpjService := cnpj.NewService(arcomClient, cnpj.NewBrasilAPIClient(), cnpj.NewCache(rdb), cnpj.NewRepo(gdb))

	prospeccaoRepo := prospeccao.NewRepo(gdb)
	buscador := prospeccao.NewBuscador(arcomClient)
	prospeccaoService := prospeccao.NewService(buscador, prospeccaoRepo)

	geoClient := geo.NewClient(cfg.GeoapifyAPIKey)

	r := apihttp.NewRouter(apihttp.Deps{
		JWTSecret:          cfg.JWTSecret,
		AuthHandler:        auth.NewHandler(authService, usuariosRepo),
		UsuariosHandler:    usuarios.NewHandler(usuariosRepo),
		CNPJHandler:        cnpj.NewHandler(cnpjService),
		ProspeccaoHandler:  prospeccao.NewHandler(prospeccaoService, prospeccaoRepo, buscador),
		PreCadastroHandler: prospeccao.NewPreCadastroHandler(prospeccaoRepo),
		GeoHandler:         geo.NewHandler(geoClient, rdb),
	})

	slog.Info("subindo servidor", "port", cfg.Port)
	if err := http.ListenAndServe("0.0.0.0:"+cfg.Port, r); err != nil {
		slog.Error("servidor encerrado", "erro", err)
	}
}
