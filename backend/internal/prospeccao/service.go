package prospeccao

import (
	"context"
	"log/slog"
)

type Service struct {
	buscador *Buscador
	repo     *Repo
}

func NewService(buscador *Buscador, repo *Repo) *Service {
	return &Service{buscador: buscador, repo: repo}
}

type BuscarInput struct {
	Cidades []CidadeFiltro `json:"cidades"`
	CNAEs   []string       `json:"cnaes"`
}

// Buscar delega ao Buscador e registra a busca no histórico (não falha a
// resposta se o log der erro — a busca em si já valeu, só perdemos o
// histórico dela).
func (s *Service) Buscar(ctx context.Context, uid string, in BuscarInput) ([]Prospect, error) {
	prospects, err := s.buscador.Buscar(ctx, in.Cidades, in.CNAEs)
	if err != nil {
		return nil, err
	}
	if uid != "" {
		if err := s.repo.RegistrarBusca(ctx, uid, in, len(prospects)); err != nil {
			slog.Error("falha ao registrar prospeccao", "erro", err)
		}
	}
	return prospects, nil
}
