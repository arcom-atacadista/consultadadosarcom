package admin

import (
	"context"
	"sort"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/atividades"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/usuarios"
)

// janelaAtividades é o mesmo teto do app antigo (250 docs mais recentes) pras
// contagens do dashboard — números "recentes", não o total histórico.
const janelaAtividades = 250

type PorUsuario struct {
	Nome         string `json:"nome"`
	Email        string `json:"email"`
	Consultas    int    `json:"consultas"`
	Prospeccoes  int    `json:"prospeccoes"`
	PDFIndicacao int    `json:"pdfIndicacao"`
}

type Dashboard struct {
	Online          int            `json:"online"`
	OnlineLista     []PresencaItem `json:"onlineLista"`
	Contas          int64          `json:"contas"`
	ContasPendentes int64          `json:"contasPendentes"`
	Logins          int            `json:"logins"`
	Consultas       int            `json:"consultas"`
	Prospeccoes     int            `json:"prospeccoes"`
	PDFs            int            `json:"pdfs"`
	PDFsCNPJs       []string       `json:"pdfsCnpjs"`
	PorUsuario      []PorUsuario   `json:"porUsuario"`
}

type Service struct {
	atividades *atividades.Repo
	usuarios   *usuarios.Repo
	presenca   *PresencaService
}

func NewService(atividadesRepo *atividades.Repo, usuariosRepo *usuarios.Repo, presenca *PresencaService) *Service {
	return &Service{atividades: atividadesRepo, usuarios: usuariosRepo, presenca: presenca}
}

func (s *Service) Dashboard(ctx context.Context) (*Dashboard, error) {
	d := &Dashboard{}

	online, err := s.presenca.Online(ctx)
	if err != nil {
		return nil, err
	}
	d.Online = len(online)
	d.OnlineLista = online

	total, pendentes, err := s.usuarios.ContarPorStatus(ctx)
	if err != nil {
		return nil, err
	}
	d.Contas = total
	d.ContasPendentes = pendentes

	ultimas, err := s.atividades.UltimasN(ctx, janelaAtividades)
	if err != nil {
		return nil, err
	}

	porUsuario := map[string]*PorUsuario{}
	cnpjsPdf := map[string]bool{}
	var cnpjsPdfOrdem []string

	for _, a := range ultimas {
		switch a.Tipo {
		case atividades.TipoLogin:
			d.Logins++
		case atividades.TipoConsulta:
			d.Consultas++
		case atividades.TipoProspeccao:
			d.Prospeccoes++
		case atividades.TipoPDFIndicacao:
			d.PDFs++
			for _, cn := range cnpjsDoPayload(a.Payload) {
				if !cnpjsPdf[cn] {
					cnpjsPdf[cn] = true
					cnpjsPdfOrdem = append(cnpjsPdfOrdem, cn)
				}
			}
		}

		chave := a.Email
		if chave == "" {
			chave = a.Nome
		}
		if chave == "" {
			continue
		}
		pu, ok := porUsuario[chave]
		if !ok {
			pu = &PorUsuario{Nome: a.Nome, Email: a.Email}
			porUsuario[chave] = pu
		}
		if pu.Nome == "" {
			pu.Nome = a.Nome
		}
		switch a.Tipo {
		case atividades.TipoConsulta:
			pu.Consultas++
		case atividades.TipoProspeccao:
			pu.Prospeccoes++
		case atividades.TipoPDFIndicacao:
			pu.PDFIndicacao++
		}
	}

	d.PDFsCNPJs = cnpjsPdfOrdem

	linhas := make([]PorUsuario, 0, len(porUsuario))
	for _, pu := range porUsuario {
		linhas = append(linhas, *pu)
	}
	sort.Slice(linhas, func(i, j int) bool {
		return (linhas[i].Consultas + linhas[i].Prospeccoes + linhas[i].PDFIndicacao) >
			(linhas[j].Consultas + linhas[j].Prospeccoes + linhas[j].PDFIndicacao)
	})
	d.PorUsuario = linhas

	return d, nil
}

func cnpjsDoPayload(payload any) []string {
	m, ok := payload.(map[string]any)
	if !ok {
		return nil
	}
	brutos, ok := m["cnpjs"].([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(brutos))
	for _, b := range brutos {
		if s, ok := b.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}
