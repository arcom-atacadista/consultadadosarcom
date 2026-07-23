package prospeccao

import "time"

type PreCadastro struct {
	ID           string    `gorm:"column:id;primaryKey" json:"id"`
	CNPJ         string    `gorm:"column:cnpj" json:"cnpj"`
	Razao        string    `gorm:"column:razao" json:"razao"`
	Endereco     string    `gorm:"column:endereco" json:"endereco"`
	Contato      string    `gorm:"column:contato" json:"contato"`
	Notas        string    `gorm:"column:notas" json:"notas"`
	Status       string    `gorm:"column:status" json:"status"`
	AutorUID     string    `gorm:"column:autor_uid" json:"autorUid"`
	CriadoEm     time.Time `gorm:"column:criado_em;autoCreateTime" json:"criadoEm"`
	AtualizadoEm time.Time `gorm:"column:atualizado_em;autoUpdateTime" json:"atualizadoEm"`
}

func (PreCadastro) TableName() string { return "pre_cadastros" }

// Filtros/Itens usam o serializer:json do próprio gorm (stdlib encoding/json
// por baixo) contra colunas jsonb — sem precisar de gorm.io/datatypes, que
// não está no allowlist (padroes/01).
type ListaProspeccao struct {
	ID       string    `gorm:"column:id;primaryKey" json:"id"`
	UID      string    `gorm:"column:uid" json:"uid"`
	Nome     string    `gorm:"column:nome" json:"nome"`
	Filtros  any       `gorm:"column:filtros;serializer:json" json:"filtros"`
	Itens    any       `gorm:"column:itens;serializer:json" json:"itens"`
	CriadoEm time.Time `gorm:"column:criado_em;autoCreateTime" json:"criadoEm"`
}

func (ListaProspeccao) TableName() string { return "listas_prospeccao" }

type ProspeccaoLog struct {
	ID       string    `gorm:"column:id;primaryKey" json:"id"`
	UID      string    `gorm:"column:uid" json:"uid"`
	Filtros  any       `gorm:"column:filtros;serializer:json" json:"filtros"`
	Total    int       `gorm:"column:total" json:"total"`
	CriadoEm time.Time `gorm:"column:criado_em;autoCreateTime" json:"criadoEm"`
}

func (ProspeccaoLog) TableName() string { return "prospeccoes" }
