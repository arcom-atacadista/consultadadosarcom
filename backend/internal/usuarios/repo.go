package usuarios

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var ErrNaoEncontrado = errors.New("usuário não encontrado")

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Criar(ctx context.Context, u *Usuario) error {
	return r.db.WithContext(ctx).Create(u).Error
}

// ContarPorStatus dá o total de contas e quantas estão pendentes de aprovação.
func (r *Repo) ContarPorStatus(ctx context.Context) (total int64, pendentes int64, err error) {
	if err = r.db.WithContext(ctx).Model(&Usuario{}).Count(&total).Error; err != nil {
		return 0, 0, err
	}
	err = r.db.WithContext(ctx).Model(&Usuario{}).Where("status = ?", StatusPendente).Count(&pendentes).Error
	return total, pendentes, err
}

func (r *Repo) BuscarPorEmail(ctx context.Context, email string) (*Usuario, error) {
	var u Usuario
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNaoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) BuscarPorID(ctx context.Context, id string) (*Usuario, error) {
	var u Usuario
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNaoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) Listar(ctx context.Context) ([]Usuario, error) {
	var us []Usuario
	err := r.db.WithContext(ctx).Order("criado_em desc").Find(&us).Error
	return us, err
}

func (r *Repo) Atualizar(ctx context.Context, id string, status *string, isAdmin *bool) error {
	updates := map[string]any{}
	if status != nil {
		updates["status"] = *status
	}
	if isAdmin != nil {
		updates["is_admin"] = *isAdmin
	}
	if len(updates) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&Usuario{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

func (r *Repo) AtualizarSenha(ctx context.Context, id, novoHash string) error {
	res := r.db.WithContext(ctx).Model(&Usuario{}).Where("id = ?", id).Update("senha_hash", novoHash)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

func (r *Repo) Deletar(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Usuario{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNaoEncontrado
	}
	return nil
}
