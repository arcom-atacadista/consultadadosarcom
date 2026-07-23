// Package http monta o router chi e agrupa as rotas sob /api.
package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter monta o router da aplicação. Novos recursos (cnpj, prospeccao,
// ia, ...) entram aqui como sub-rotas conforme as fases da migração avançam.
func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)

	r.Route("/api", func(api chi.Router) {
		api.Get("/health", Health)
	})

	return r
}
