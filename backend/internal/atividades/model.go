package atividades

import "time"

// Tipos de atividade — espelham o app antigo (ATIVIDADE_META), menos
// "razao_social" que já era código morto por lá (nenhum call site registrava).
const (
	TipoContaCriada  = "conta_criada"
	TipoLogin        = "login"
	TipoConsulta     = "consulta"
	TipoProspeccao   = "prospeccao"
	TipoPreCadastro  = "precadastro"
	TipoPDFIndicacao = "pdf_indicacao"
)

// TiposApagaveis são os únicos tipos que a limpeza de logs antigos remove —
// só ruído de acesso, nunca dado de CNPJ (mesma regra do app antigo).
var TiposApagaveis = []string{TipoLogin, TipoContaCriada}

type Atividade struct {
	ID       string    `gorm:"column:id;primaryKey" json:"id"`
	Tipo     string    `gorm:"column:tipo" json:"tipo"`
	UID      string    `gorm:"column:uid" json:"uid"`
	Nome     string    `gorm:"column:nome" json:"nome"`
	Email    string    `gorm:"column:email" json:"email"`
	Detalhe  string    `gorm:"column:detalhe" json:"detalhe"`
	Payload  any       `gorm:"column:payload;serializer:json" json:"payload,omitempty"`
	CriadoEm time.Time `gorm:"column:criado_em;autoCreateTime" json:"criadoEm"`
}

func (Atividade) TableName() string { return "atividades_log" }
