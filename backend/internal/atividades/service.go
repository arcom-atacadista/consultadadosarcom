package atividades

import (
	"context"
	"log/slog"
)

// Service registra atividades em modo best-effort: uma falha aqui nunca pode
// derrubar a operação principal (consulta, prospecção, login...) — mesma
// postura do app antigo (registrarAtividade tinha um try/catch que só
// logava no console e seguia).
type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) Registrar(ctx context.Context, uid, nome, email, tipo, detalhe string, payload any) {
	if s == nil {
		return
	}
	err := s.repo.Registrar(ctx, &Atividade{
		Tipo: tipo, UID: uid, Nome: nome, Email: email, Detalhe: detalhe, Payload: payload,
	})
	if err != nil {
		slog.Warn("falha ao registrar atividade", "tipo", tipo, "erro", err)
	}
}

func (s *Service) Repo() *Repo { return s.repo }
