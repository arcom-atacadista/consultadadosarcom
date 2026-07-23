package atividades

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Registrar(ctx context.Context, a *Atividade) error {
	a.ID = uuid.NewString()
	return r.db.WithContext(ctx).Create(a).Error
}

// UltimasN busca as N atividades mais recentes — mesma janela usada pelo
// app antigo (250 docs) pras contagens do dashboard, pra não ler a tabela
// inteira a cada refresh.
func (r *Repo) UltimasN(ctx context.Context, n int) ([]Atividade, error) {
	var lista []Atividade
	err := r.db.WithContext(ctx).Order("criado_em DESC").Limit(n).Find(&lista).Error
	return lista, err
}

func (r *Repo) LimparAntigos(ctx context.Context, dias int, tipos []string) (int64, error) {
	corte := time.Now().AddDate(0, 0, -dias)
	res := r.db.WithContext(ctx).
		Where("criado_em < ? AND tipo IN ?", corte, tipos).
		Delete(&Atividade{})
	return res.RowsAffected, res.Error
}
