package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/auth"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/cnpj"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/config"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/db"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/enriquecimento"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/geo"
	apihttp "github.com/arcom-atacadista/consultadadosarcom/backend/internal/http"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/ia"
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

	cnpjRepo := cnpj.NewRepo(gdb)
	cnpjService := cnpj.NewService(arcomClient, cnpj.NewBrasilAPIClient(), cnpj.NewCache(rdb), cnpjRepo)

	prospeccaoRepo := prospeccao.NewRepo(gdb)
	buscador := prospeccao.NewBuscador(arcomClient)
	prospeccaoService := prospeccao.NewService(buscador, prospeccaoRepo)

	geoClient := geo.NewClient(cfg.GeoapifyAPIKey)

	groqClient := ia.NewGroqClient(cfg.GroqAPIKey, cfg.GroqAPIURL)
	tavilyClient := ia.NewTavilyClient(cfg.TavilyAPIKey, cfg.TavilyAPIURL)
	insightService := ia.NewInsightService(groqClient, tavilyClient)
	rankingService := ia.NewRankingService(groqClient)
	chatService := ia.NewChatService(groqClient, tavilyClient, montarEstatisticasFn(usuariosRepo, cnpjRepo, prospeccaoRepo))

	trace360Client := enriquecimento.NewTrace360Client(cfg.Trace360BaseURL, cfg.Trace360APIKey)
	enriquecimentoRepo := enriquecimento.NewRepo(gdb)
	enriquecimentoService := enriquecimento.NewService(trace360Client, enriquecimentoRepo)

	r := apihttp.NewRouter(apihttp.Deps{
		JWTSecret:             cfg.JWTSecret,
		AuthHandler:           auth.NewHandler(authService, usuariosRepo),
		UsuariosHandler:       usuarios.NewHandler(usuariosRepo),
		CNPJHandler:           cnpj.NewHandler(cnpjService),
		ProspeccaoHandler:     prospeccao.NewHandler(prospeccaoService, prospeccaoRepo, buscador),
		PreCadastroHandler:    prospeccao.NewPreCadastroHandler(prospeccaoRepo),
		GeoHandler:            geo.NewHandler(geoClient, rdb),
		IAHandler:             ia.NewHandler(insightService, rankingService, chatService),
		EnriquecimentoHandler: enriquecimento.NewHandler(enriquecimentoService, trace360Client),
	})

	slog.Info("subindo servidor", "port", cfg.Port)
	if err := http.ListenAndServe("0.0.0.0:"+cfg.Port, r); err != nil {
		slog.Error("servidor encerrado", "erro", err)
	}
}

// montarEstatisticasFn compõe a ferramenta "estatisticas_do_site" do chat com
// números reais (contas, consultas, prospecções). Atividade por usuário e
// presença online chegam na Fase 7 (tabela atividades_log + Redis) — até lá,
// reporta só o que já existe, nunca inventa.
func montarEstatisticasFn(usuariosRepo *usuarios.Repo, cnpjRepo *cnpj.Repo, prospeccaoRepo *prospeccao.Repo) ia.EstatisticasFn {
	return func(ctx context.Context) (string, error) {
		total, pendentes, err := usuariosRepo.ContarPorStatus(ctx)
		if err != nil {
			return "", err
		}
		consultas, err := cnpjRepo.ContarConsultas(ctx)
		if err != nil {
			return "", err
		}
		buscas, prospects, err := prospeccaoRepo.ContarBuscas(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(
			"Estatísticas do site (agora — %s):\n- Contas: %d no total, %d pendente(s) de aprovação\n- Consultas de CNPJ realizadas: %d\n- Buscas de prospecção: %d (%d empresa(s) encontradas no total)",
			time.Now().Format("02/01/2006 15:04"), total, pendentes, consultas, buscas, prospects,
		), nil
	}
}
