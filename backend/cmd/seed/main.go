// Comando seed lê o export do Firestore (docs/migracao/04 §4) e insere os
// dados vivos no Postgres novo, pro cutover da Fase 7. Não roda sozinho: quem
// faz o cutover precisa antes gerar o export com o Firebase Admin SDK ou
// `gcloud firestore export`, converter pro formato JSON esperado aqui
// (um array por coleção — usuarios/preCadastros/listasProspeccao) e então
// rodar:
//
//	go run ./cmd/seed -arquivo export.json
//
// Senhas do Firebase Auth não migram (hash proprietário, não exportável) —
// os usuários migrados entram sem senha (usuarios.senha_hash vazio) e
// precisam redefinir a senha no primeiro acesso (ver docs/migracao/10 §1.4).
// Histórico de logs (consultas_log/atividades_log) deliberadamente NÃO migra
// — decisão já tomada em docs/migracao/10 §1.5, começa limpo no Postgres.
package main

import (
	"encoding/json"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/config"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/db"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/prospeccao"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/usuarios"
)

// exportUsuario/exportPreCadastro/exportListaProspeccao espelham os campos
// das coleções do Firestore (docs/migracao/04 §1) — ajuste os nomes de campo
// aqui se o export real vier com chaves diferentes.
type exportUsuario struct {
	UID      string `json:"uid"`
	Nome     string `json:"nome"`
	Email    string `json:"email"`
	Status   string `json:"status"`
	IsAdmin  bool   `json:"isAdmin"`
	CriadoEm string `json:"criadoEm"`
}

type exportPreCadastro struct {
	CNPJ     string `json:"cnpj"`
	Razao    string `json:"razao"`
	Endereco string `json:"endereco"`
	Contato  string `json:"contato"`
	Notas    string `json:"notas"`
	Status   string `json:"status"`
	AutorUID string `json:"autorUid"`
}

type exportListaProspeccao struct {
	UID      string `json:"uid"`
	Nome     string `json:"nome"`
	Assessor string `json:"assessor"`
	Filtros  any    `json:"filtros"`
	Itens    any    `json:"itens"`
}

type exportado struct {
	Usuarios         []exportUsuario         `json:"usuarios"`
	PreCadastros     []exportPreCadastro     `json:"preCadastros"`
	ListasProspeccao []exportListaProspeccao `json:"listasProspeccao"`
}

func main() {
	arquivo := flag.String("arquivo", "", "caminho do JSON exportado do Firestore")
	flag.Parse()
	if *arquivo == "" {
		slog.Error("uso: go run ./cmd/seed -arquivo export.json")
		os.Exit(1)
	}

	raw, err := os.ReadFile(*arquivo)
	if err != nil {
		slog.Error("falha ao ler o arquivo de export", "erro", err)
		os.Exit(1)
	}
	var dados exportado
	if err := json.Unmarshal(raw, &dados); err != nil {
		slog.Error("falha ao decodificar o JSON de export", "erro", err)
		os.Exit(1)
	}

	cfg := config.Load()
	gdb, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("falha ao conectar no postgres", "erro", err)
		os.Exit(1)
	}

	// uidAntigoParaNovo mapeia o uid do Firebase Auth (usado nas referências
	// de preCadastros.autorUid e listasProspeccao.uid) pro novo id gerado no
	// Postgres — os ids não podem ser reaproveitados 1:1 (usuarios.id aqui é
	// gerado por gen_random_uuid(), não o uid do Firebase).
	uidAntigoParaNovo := map[string]string{}

	totalUsuarios := 0
	for _, u := range dados.Usuarios {
		if u.Status != usuarios.StatusAprovado {
			// docs/migracao/10 §1.4: só os aprovados migram — contas
			// pendentes/negadas do app antigo o usuário refaz o cadastro.
			continue
		}
		novo := usuarios.Usuario{
			ID: uuid.NewString(), Email: u.Email, Nome: u.Nome,
			Status: usuarios.StatusAprovado, IsAdmin: u.IsAdmin,
			// SenhaHash fica vazio de propósito — ver comentário no topo do arquivo.
		}
		if err := gdb.Create(&novo).Error; err != nil {
			slog.Warn("falha ao inserir usuário", "email", u.Email, "erro", err)
			continue
		}
		uidAntigoParaNovo[u.UID] = novo.ID
		totalUsuarios++
	}
	slog.Info("usuários migrados", "total", totalUsuarios, "descartados_nao_aprovados", len(dados.Usuarios)-totalUsuarios)

	totalPreCadastros := 0
	for _, p := range dados.PreCadastros {
		novo := prospeccao.PreCadastro{
			ID: uuid.NewString(), CNPJ: p.CNPJ, Razao: p.Razao, Endereco: p.Endereco,
			Contato: p.Contato, Notas: p.Notas, Status: p.Status,
			AutorUID: uidAntigoParaNovo[p.AutorUID],
		}
		if err := gdb.Create(&novo).Error; err != nil {
			slog.Warn("falha ao inserir pré-cadastro", "cnpj", p.CNPJ, "erro", err)
			continue
		}
		totalPreCadastros++
	}
	slog.Info("pré-cadastros migrados", "total", totalPreCadastros)

	totalListas := 0
	for _, l := range dados.ListasProspeccao {
		uidNovo, ok := uidAntigoParaNovo[l.UID]
		if !ok {
			slog.Warn("lista ignorada: uid do dono não foi migrado (conta não aprovada?)", "lista", l.Nome)
			continue
		}
		novo := prospeccao.ListaProspeccao{
			ID: uuid.NewString(), UID: uidNovo, Nome: l.Nome, Assessor: l.Assessor,
			Filtros: l.Filtros, Itens: l.Itens,
		}
		if err := gdb.Create(&novo).Error; err != nil {
			slog.Warn("falha ao inserir lista de prospecção", "nome", l.Nome, "erro", err)
			continue
		}
		totalListas++
	}
	slog.Info("listas de prospecção migradas", "total", totalListas)

	slog.Info("seed concluído", "quando", time.Now().Format(time.RFC3339))
}
