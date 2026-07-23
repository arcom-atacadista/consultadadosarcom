package cnpj

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConsultaLogEntry é uma linha de consultas_log — usada pro contador de uso
// mensal por usuário (dashboard admin, Fase 7).
type ConsultaLogEntry struct {
	ID       string    `gorm:"column:id;primaryKey"`
	UID      string    `gorm:"column:uid"`
	CNPJ     string    `gorm:"column:cnpj"`
	Fonte    string    `gorm:"column:fonte"`
	CriadoEm time.Time `gorm:"column:criado_em;autoCreateTime"`
}

func (ConsultaLogEntry) TableName() string { return "consultas_log" }

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

// RegistrarConsultas grava uma linha por CNPJ efetivamente buscado na fonte
// externa (cache hit não conta — não gastou chamada de API).
func (r *Repo) RegistrarConsultas(ctx context.Context, uid string, cnpjs []string, fonte string) error {
	if len(cnpjs) == 0 {
		return nil
	}
	entradas := make([]ConsultaLogEntry, 0, len(cnpjs))
	for _, c := range cnpjs {
		entradas = append(entradas, ConsultaLogEntry{ID: uuid.NewString(), UID: uid, CNPJ: c, Fonte: fonte})
	}
	return r.db.WithContext(ctx).Create(&entradas).Error
}
