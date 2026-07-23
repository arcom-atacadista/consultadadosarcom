package prospeccao

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/cnpj"
)

// ── "Conversão da prospecção" (Fase 7) ──────────────────────────────────────
// Retrato de cada lista salva pra um assessor: de tudo que foi prospectado,
// quanto virou Cliente Arcom. Reproduz fielmente a Feature B do app antigo
// (carregarConversao/renderConversao/verificarConversaoTodas) — a Feature A
// (upload de planilha cruzando com toda prospecção já feita) não migrou:
// era código morto em produção (duas funções `renderConversao` no mesmo
// escopo, a segunda sempre vencia, então o upload nunca funcionava de
// verdade) e a decisão foi não portar uma função que já não fazia nada.

type filtrosLista struct {
	Cidades     []CidadeFiltro `json:"cidades"`
	RamosLabels []string       `json:"ramosLabels"`
}

func parseFiltros(filtros any) filtrosLista {
	var f filtrosLista
	b, err := json.Marshal(filtros)
	if err != nil {
		return f
	}
	_ = json.Unmarshal(b, &f)
	return f
}

func cidadeDaLista(l ListaProspeccao) string {
	f := parseFiltros(l.Filtros)
	partes := make([]string, 0, len(f.Cidades))
	for _, c := range f.Cidades {
		partes = append(partes, c.Cidade+"/"+c.UF)
	}
	return strings.Join(partes, ", ")
}

func ramoDaLista(l ListaProspeccao) string {
	f := parseFiltros(l.Filtros)
	return strings.Join(f.RamosLabels, ", ")
}

func itensParaCNPJs(itens any) []string {
	brutos, ok := itens.([]any)
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

type ItemRankingConversao struct {
	Nome       string `json:"nome"`
	Total      int    `json:"total"`
	Convertido int    `json:"convertido"`
	Verificado bool   `json:"verificado"`
}

type ListaConversao struct {
	ID          string `json:"id"`
	Nome        string `json:"nome"`
	NomeUsuario string `json:"nomeUsuario"`
	Assessor    string `json:"assessor"`
	Cidade      string `json:"cidade"`
	CriadoEm    string `json:"criadoEm"`
	Total       int    `json:"total"`
	Convertidos *int   `json:"convertidos"`
}

type RelatorioConversao struct {
	TotalListas       int                    `json:"totalListas"`
	TotalEmpresas     int                    `json:"totalEmpresas"`
	TotalConvertidas  int                    `json:"totalConvertidas"`
	TaxaConversao     float64                `json:"taxaConversao"`
	AlgumaVerificada  bool                   `json:"algumaVerificada"`
	PorAssessor       []ItemRankingConversao `json:"porAssessor"`
	PorQuemProspectou []ItemRankingConversao `json:"porQuemProspectou"`
	PorCidade         []ItemRankingConversao `json:"porCidade"`
	Listas            []ListaConversao       `json:"listas"`
}

type ConversaoService struct {
	repo  *Repo
	arcom *cnpj.ArcomClient
}

func NewConversaoService(repo *Repo, arcom *cnpj.ArcomClient) *ConversaoService {
	return &ConversaoService{repo: repo, arcom: arcom}
}

// Relatorio reproduz renderConversao(listas) do app antigo: agregações por
// assessor / por quem prospectou / por cidade, e a tabela de listas.
func (s *ConversaoService) Relatorio(ctx context.Context, dias int) (*RelatorioConversao, error) {
	listas, err := s.repo.ListarListasPeriodo(ctx, dias)
	if err != nil {
		return nil, err
	}

	rel := &RelatorioConversao{TotalListas: len(listas)}
	agAssessor := map[string]*ItemRankingConversao{}
	agProspector := map[string]*ItemRankingConversao{}
	agCidade := map[string]*ItemRankingConversao{}

	ag := func(mapa map[string]*ItemRankingConversao, nomeCru string, total int, conv *int) {
		nome := strings.TrimSpace(nomeCru)
		if nome == "" {
			nome = "—"
		}
		chave := semAcento(nome)
		item, ok := mapa[chave]
		if !ok {
			item = &ItemRankingConversao{Nome: nome}
			mapa[chave] = item
		}
		item.Total += total
		if conv != nil {
			item.Convertido += *conv
			item.Verificado = true
		}
	}

	rel.Listas = make([]ListaConversao, 0, len(listas))
	for _, l := range listas {
		total := len(itensParaCNPJs(l.Itens))
		if l.TotalEmpresas != nil {
			total = *l.TotalEmpresas
		}
		rel.TotalEmpresas += total
		if l.Convertidos != nil {
			rel.TotalConvertidas += *l.Convertidos
			rel.AlgumaVerificada = true
		}

		ag(agAssessor, l.Assessor, total, l.Convertidos)
		prospector := l.NomeUsuario
		if prospector == "" {
			prospector = l.Email
		}
		ag(agProspector, prospector, total, l.Convertidos)
		ag(agCidade, cidadeDaLista(l), total, l.Convertidos)

		rel.Listas = append(rel.Listas, ListaConversao{
			ID: l.ID, Nome: l.Nome, NomeUsuario: prospector, Assessor: l.Assessor,
			Cidade: cidadeDaLista(l), CriadoEm: l.CriadoEm.Format("2006-01-02T15:04:05Z07:00"),
			Total: total, Convertidos: l.Convertidos,
		})
	}
	if rel.TotalEmpresas > 0 {
		rel.TaxaConversao = float64(rel.TotalConvertidas) / float64(rel.TotalEmpresas) * 100
	}
	rel.PorAssessor = ordenarRanking(agAssessor)
	rel.PorQuemProspectou = ordenarRanking(agProspector)
	rel.PorCidade = ordenarRanking(agCidade)
	return rel, nil
}

func ordenarRanking(mapa map[string]*ItemRankingConversao) []ItemRankingConversao {
	out := make([]ItemRankingConversao, 0, len(mapa))
	for _, v := range mapa {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Convertido != out[j].Convertido {
			return out[i].Convertido > out[j].Convertido
		}
		return out[i].Total > out[j].Total
	})
	return out
}

// Verificar reconsulta a API Arcom (com cliente:true) pra todos os CNPJs das
// listas do período e persiste convertidos/totalEmpresas/verificadoEm em
// cada lista — igual ao verificarConversaoTodas do app antigo.
func (s *ConversaoService) Verificar(ctx context.Context, dias int) (*RelatorioConversao, error) {
	listas, err := s.repo.ListarListasPeriodo(ctx, dias)
	if err != nil {
		return nil, err
	}

	unicos := map[string]bool{}
	var todosCNPJs []string
	for _, l := range listas {
		for _, c := range itensParaCNPJs(l.Itens) {
			if !unicos[c] {
				unicos[c] = true
				todosCNPJs = append(todosCNPJs, c)
			}
		}
	}

	jaCliente := map[string]bool{}
	if len(todosCNPJs) > 0 {
		resultados, err := s.arcom.ConsultarLote(ctx, todosCNPJs)
		if err != nil {
			return nil, err
		}
		for _, item := range resultados {
			if item.CNPJ != "" {
				jaCliente[item.CNPJ] = item.JaCliente
			}
		}
	}

	for _, l := range listas {
		cnpjs := itensParaCNPJs(l.Itens)
		conv := 0
		for _, c := range cnpjs {
			if jaCliente[c] {
				conv++
			}
		}
		if err := s.repo.AtualizarConversao(ctx, l.ID, conv, len(cnpjs)); err != nil {
			return nil, err
		}
	}

	return s.Relatorio(ctx, dias)
}
