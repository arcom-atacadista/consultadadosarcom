package prospeccao

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNaoEncontrado = errors.New("registro não encontrado")

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

// ── Pré-cadastros (compartilhados entre todos os aprovados) ────────────────

func (r *Repo) CriarPreCadastro(ctx context.Context, p *PreCadastro) error {
	p.ID = uuid.NewString()
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *Repo) ListarPreCadastros(ctx context.Context) ([]PreCadastro, error) {
	var out []PreCadastro
	err := r.db.WithContext(ctx).Order("criado_em desc").Find(&out).Error
	return out, err
}

func (r *Repo) AtualizarPreCadastro(ctx context.Context, id, status, notas string) error {
	updates := map[string]any{}
	if status != "" {
		updates["status"] = status
	}
	if notas != "" {
		updates["notas"] = notas
	}
	if len(updates) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&PreCadastro{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

func (r *Repo) DeletarPreCadastro(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&PreCadastro{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

// ── Listas de prospecção salvas (por usuário) ───────────────────────────────

func (r *Repo) CriarLista(ctx context.Context, l *ListaProspeccao) error {
	l.ID = uuid.NewString()
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *Repo) ListarListas(ctx context.Context, uid string, admin bool) ([]ListaProspeccao, error) {
	var out []ListaProspeccao
	q := r.db.WithContext(ctx).Order("criado_em desc")
	if !admin {
		q = q.Where("uid = ?", uid)
	}
	err := q.Find(&out).Error
	return out, err
}

func (r *Repo) BuscarLista(ctx context.Context, id string) (*ListaProspeccao, error) {
	var l ListaProspeccao
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&l).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNaoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// maxListasConversao é o mesmo teto do app antigo (500 listas mais recentes)
// pro relatório de Conversão.
const maxListasConversao = 500

// ListarListasPeriodo busca as listas mais recentes (teto de 500, igual ao
// app antigo) e filtra por data de criação quando um corte é informado — a
// mesma lógica de "Tudo / Últimos 30 dias / Últimos 90 dias" da Conversão.
func (r *Repo) ListarListasPeriodo(ctx context.Context, dias int) ([]ListaProspeccao, error) {
	q := r.db.WithContext(ctx).Order("criado_em desc").Limit(maxListasConversao)
	if dias > 0 {
		corte := time.Now().AddDate(0, 0, -dias)
		q = q.Where("criado_em >= ?", corte)
	}
	var out []ListaProspeccao
	err := q.Find(&out).Error
	return out, err
}

// AtualizarConversao persiste o resultado de "Verificar conversão agora"
// numa lista — igual ao listasCol.doc(l.id).update(...) do app antigo.
func (r *Repo) AtualizarConversao(ctx context.Context, id string, convertidos, totalEmpresas int) error {
	agora := time.Now()
	return r.db.WithContext(ctx).Model(&ListaProspeccao{}).Where("id = ?", id).Updates(map[string]any{
		"convertidos":    convertidos,
		"total_empresas": totalEmpresas,
		"verificado_em":  agora,
	}).Error
}

func (r *Repo) DeletarLista(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&ListaProspeccao{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

// ── Histórico de buscas (prospeccoes) ───────────────────────────────────────

func (r *Repo) RegistrarBusca(ctx context.Context, uid string, filtros any, total int) error {
	log := ProspeccaoLog{ID: uuid.NewString(), UID: uid, Filtros: filtros, Total: total}
	return r.db.WithContext(ctx).Create(&log).Error
}

// ContarBuscas dá o total de buscas de prospecção feitas e a soma de
// prospects encontrados em todas elas (usado em estatisticas_do_site).
func (r *Repo) ContarBuscas(ctx context.Context) (buscas int64, totalProspects int64, err error) {
	if err = r.db.WithContext(ctx).Model(&ProspeccaoLog{}).Count(&buscas).Error; err != nil {
		return 0, 0, err
	}
	var soma *int64
	if err = r.db.WithContext(ctx).Model(&ProspeccaoLog{}).Select("SUM(total)").Scan(&soma).Error; err != nil {
		return 0, 0, err
	}
	if soma != nil {
		totalProspects = *soma
	}
	return buscas, totalProspects, nil
}
