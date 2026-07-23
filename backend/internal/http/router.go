// Package http monta o router chi e agrupa as rotas sob /api.
package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/auth"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/usuarios"
)

// Deps são as dependências já montadas em main.go que o router precisa
// para expor as rotas autenticadas/administrativas.
type Deps struct {
	JWTSecret       string
	AuthHandler     *auth.Handler
	UsuariosHandler *usuarios.Handler
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
	})

	return r
}
