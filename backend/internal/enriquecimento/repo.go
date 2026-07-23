package enriquecimento

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNaoEncontrado = errors.New("enriquecimento não encontrado")

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

// Upsert grava (ou atualiza) a posse: um usuário só vê os CNPJs que ele
// mesmo enviou — a chave da Trace360 é única pro site inteiro, então a posse
// vira responsabilidade do backend (antes ficava no localStorage do app antigo).
func (r *Repo) Upsert(ctx context.Context, e *Enriquecimento) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uid"}, {Name: "cnpj"}},
			DoUpdates: clause.AssignmentColumns([]string{"cliente_id", "status"}),
		}).
		Create(e).Error
}

func (r *Repo) ListarPorUsuario(ctx context.Context, uid string, limite int) ([]Enriquecimento, error) {
	var out []Enriquecimento
	err := r.db.WithContext(ctx).
		Where("uid = ?", uid).
		Order("criado_em desc").
		Limit(limite).
		Find(&out).Error
	return out, err
}

func (r *Repo) BuscarPorClienteID(ctx context.Context, clienteID string) (*Enriquecimento, error) {
	var e Enriquecimento
	err := r.db.WithContext(ctx).Where("cliente_id = ?", clienteID).First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNaoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repo) AtualizarStatus(ctx context.Context, id, status, razaoSocial string) error {
	updates := map[string]any{"status": status}
	if razaoSocial != "" {
		updates["razao_social"] = razaoSocial
	}
	return r.db.WithContext(ctx).Model(&Enriquecimento{}).Where("id = ?", id).Updates(updates).Error
}
