// Package http monta o router chi e agrupa as rotas sob /api.
package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/admin"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/atividades"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/auth"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/cnpj"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/enriquecimento"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/geo"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/ia"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/prospeccao"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/usuarios"
)

// Deps são as dependências já montadas em main.go que o router precisa
// para expor as rotas autenticadas/administrativas.
type Deps struct {
	JWTSecret             string
	AuthHandler           *auth.Handler
	UsuariosHandler       *usuarios.Handler
	CNPJHandler           *cnpj.Handler
	ProspeccaoHandler     *prospeccao.Handler
	PreCadastroHandler    *prospeccao.PreCadastroHandler
	ConversaoHandler      *prospeccao.ConversaoHandler
	GeoHandler            *geo.Handler
	IAHandler             *ia.Handler
	EnriquecimentoHandler *enriquecimento.Handler
	AtividadesHandler     *atividades.Handler
	AdminHandler          *admin.Handler
	PresencaHandler       *admin.PresencaHandler
}

// NewRouter monta o router da aplicação. Novos recursos (cnpj, prospeccao,
// ia, ...) entram aqui como sub-rotas conforme as fases da migração avançam.
func NewRouter(deps Deps) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)

	requireAuth := auth.RequireAuth(deps.JWTSecret)

	r.Route("/api", func(api chi.Router) {
		api.Get("/health", Health)

		api.Route("/auth", func(a chi.Router) {
			a.Post("/register", deps.AuthHandler.Register)
			a.Post("/login", deps.AuthHandler.Login)
			a.With(requireAuth).Get("/me", deps.AuthHandler.Me)
			a.With(requireAuth).Post("/senha", deps.AuthHandler.TrocarSenha)
		})

		api.Route("/usuarios", func(u chi.Router) {
			u.Use(requireAuth, auth.RequireAdmin)
			u.Mount("/", deps.UsuariosHandler.Routes())
		})

		api.Route("/cnpj", func(c chi.Router) {
			c.Use(requireAuth, auth.RequireAprovado)
			c.Post("/consultar", deps.CNPJHandler.Consultar)
		})

		api.Route("/prospeccao", func(p chi.Router) {
			p.Use(requireAuth, auth.RequireAprovado)
			p.Mount("/", deps.ProspeccaoHandler.Routes())
		})

		api.Route("/precadastros", func(p chi.Router) {
			p.Use(requireAuth, auth.RequireAprovado)
			p.Mount("/", deps.PreCadastroHandler.Routes())
		})

		api.Route("/geo", func(g chi.Router) {
			g.Use(requireAuth, auth.RequireAprovado)
			g.Get("/geocode", deps.GeoHandler.Geocode)
		})

		api.Route("/ia", func(i chi.Router) {
			i.Use(requireAuth, auth.RequireAprovado)
			i.Post("/insight", deps.IAHandler.Insight)
			i.Post("/ranking", deps.IAHandler.Ranking)
			i.Post("/chat", deps.IAHandler.Chat)
		})

		api.Route("/enriquecimento", func(e chi.Router) {
			e.Use(requireAuth, auth.RequireAprovado)
			e.Mount("/", deps.EnriquecimentoHandler.Routes())
		})

		api.Route("/atividades", func(a chi.Router) {
			a.Use(requireAuth, auth.RequireAprovado)
			a.Mount("/", deps.AtividadesHandler.Routes())
		})

		api.Route("/presenca", func(p chi.Router) {
			p.Use(requireAuth, auth.RequireAprovado)
			p.Post("/", deps.PresencaHandler.Heartbeat)
			p.Delete("/", deps.PresencaHandler.Remover)
		})

		api.Route("/conversao", func(c chi.Router) {
			c.Use(requireAuth, auth.RequireAdmin)
			c.Mount("/", deps.ConversaoHandler.Routes())
		})

		api.Route("/admin", func(a chi.Router) {
			a.Use(requireAuth, auth.RequireAdmin)
			a.Mount("/", deps.AdminHandler.Routes())
		})
	})

	return r
}
