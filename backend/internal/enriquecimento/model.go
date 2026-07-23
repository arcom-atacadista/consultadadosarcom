package enriquecimento

import "time"

// Estados terminais da Trace360 — quando parar de fazer polling.
var StatusTerminais = map[string]bool{
	"concluido": true, "nao_enriquecivel": true, "erro": true, "cancelado": true,
}

// EmAndamento espelha os estados do app antigo que ainda pedem atualização.
var StatusEmAndamento = map[string]bool{
	"pendente": true, "enfileirado": true, "processando": true, "ja_ativo": true, "ambiguo_pausado": true,
}

type Enriquecimento struct {
	ID           string    `gorm:"column:id;primaryKey" json:"id"`
	UID          string    `gorm:"column:uid" json:"uid"`
	CNPJ         string    `gorm:"column:cnpj" json:"cnpj"`
	ClienteID    string    `gorm:"column:cliente_id" json:"clienteId"`
	Status       string    `gorm:"column:status" json:"status"`
	RazaoSocial  string    `gorm:"column:razao_social" json:"razaoSocial"`
	CriadoEm     time.Time `gorm:"column:criado_em;autoCreateTime" json:"criadoEm"`
	AtualizadoEm time.Time `gorm:"column:atualizado_em;autoUpdateTime" json:"atualizadoEm"`
}

func (Enriquecimento) TableName() string { return "enriquecimentos" }
