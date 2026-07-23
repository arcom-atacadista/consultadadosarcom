package usuarios

import "time"

const (
	StatusPendente = "pendente"
	StatusAprovado = "aprovado"
)

type Usuario struct {
	ID        string    `gorm:"column:id;primaryKey" json:"id"`
	Email     string    `gorm:"column:email" json:"email"`
	SenhaHash string    `gorm:"column:senha_hash" json:"-"`
	Nome      string    `gorm:"column:nome" json:"nome"`
	Status    string    `gorm:"column:status" json:"status"`
	IsAdmin   bool      `gorm:"column:is_admin" json:"isAdmin"`
	CriadoEm  time.Time `gorm:"column:criado_em;autoCreateTime" json:"criadoEm"`
}

func (Usuario) TableName() string { return "usuarios" }
