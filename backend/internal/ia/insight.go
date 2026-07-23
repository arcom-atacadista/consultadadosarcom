package ia

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/cnpj"
)

type ContatosInsight struct {
	TelefoneOficial  string            `json:"telefoneOficial"`
	TelefoneOficial2 string            `json:"telefoneOficial2"`
	TelefoneWeb      string            `json:"telefoneWeb"`
	Telefone         string            `json:"telefone"`
	Site             string            `json:"site"`
	Instagram        string            `json:"instagram"`
	LinkedIn         string            `json:"linkedin"`
	Email            string            `json:"email"`
	Fontes           []TavilyResultado `json:"fontes"`
}

type Insight struct {
	Resumo                string          `json:"resumo"`
	PontosFortes          []string        `json:"pontosFortes"`
	SinaisAtencao         []string        `json:"sinaisAtencao"`
	AbordagemSugerida     string          `json:"abordagemSugerida"`
	PerguntasQualificacao []string        `json:"perguntasQualificacao"`
	NivelConfianca        string          `json:"nivelConfianca"`
	BuscaWebRealizada     bool            `json:"buscaWebRealizada"`
	Contatos              ContatosInsight `json:"contatos"`
}

type respostaGroqInsight struct {
	Resumo                string   `json:"resumo"`
	PontosFortes          []string `json:"pontos_fortes"`
	SinaisAtencao         []string `json:"sinais_atencao"`
	AbordagemSugerida     string   `json:"abordagem_sugerida"`
	PerguntasQualificacao []string `json:"perguntas_qualificacao"`
	NivelConfianca        string   `json:"nivel_confianca"`
	ContatosEncontrados   struct {
		Telefone  *string `json:"telefone"`
		Site      *string `json:"site"`
		Instagram *string `json:"instagram"`
		LinkedIn  *string `json:"linkedin"`
		Email     *string `json:"email"`
		Fontes    []int   `json:"fontes"`
	} `json:"contatos_encontrados"`
}

type InsightService struct {
	groq   *GroqClient
	tavily *TavilyClient
}

func NewInsightService(groq *GroqClient, tavily *TavilyClient) *InsightService {
	return &InsightService{groq: groq, tavily: tavily}
}

// Gerar reproduz o fluxo do app antigo: 1) Tavily busca de verdade na web
// (site, Instagram, LinkedIn, telefone); 2) Groq lê os trechos e monta o
// insight comercial, citando de onde tirou cada contato — nunca inventa.
func (s *InsightService) Gerar(ctx context.Context, e cnpj.Empresa) (*Insight, error) {
	nome := e.Razao
	if e.NomeFantasia != "" && e.NomeFantasia != "—" {
		nome = e.NomeFantasia
	}
	query := fmt.Sprintf("%s %s %s telefone contato site instagram linkedin", nome, e.Municipio, e.UF)

	var resultadosWeb []TavilyResultado
	if s.tavily.apiKey != "" {
		_, res, err := s.tavily.Buscar(ctx, query, "advanced", 8, false)
		if err == nil {
			for i := range res {
				if len(res[i].Content) > 500 {
					res[i].Content = res[i].Content[:500]
				}
			}
			resultadosWeb = res
		}
	}

	blocoWeb := "(nenhum resultado de busca disponível — Tavily não configurada ou sem resultados)"
	if len(resultadosWeb) > 0 {
		partes := make([]string, len(resultadosWeb))
		for i, r := range resultadosWeb {
			partes[i] = fmt.Sprintf("[%d] %s\nURL: %s\nTrecho: %s", i+1, r.Title, r.URL, r.Content)
		}
		blocoWeb = strings.Join(partes, "\n\n")
	}

	telOficial := semTraco(e.Telefone)

	socios := make([]string, 0, len(e.Socios))
	for _, s := range e.Socios {
		socios = append(socios, fmt.Sprintf("%s (%s)", s.NomeSocio, s.QualificacaoSocio))
	}
	sociosTxt := "não informado"
	if len(socios) > 0 {
		sociosTxt = strings.Join(socios, "; ")
	}

	prompt := fmt.Sprintf(`Você é um analista comercial. Use os dados públicos da Receita Federal E os trechos de busca na internet abaixo para montar um insight comercial de abordagem B2B.

Dados da empresa (Receita Federal):
- Razão Social: %s
- Nome Fantasia: %s
- Situação Cadastral: %s
- Porte: %s
- Natureza Jurídica: %s
- Atividade Principal: %s
- Data de Abertura: %s
- Capital Social: %s
- Município/UF: %s/%s
- Telefone OFICIAL (Receita Federal): %s
- Quadro Societário: %s

Resultados de busca na internet sobre esta empresa:
%s

REGRAS IMPORTANTES:
- OBJETIVO PRINCIPAL: ENCONTRE o telefone de contato da empresa NOS TRECHOS DE BUSCA (site oficial, Google, Google Maps, Instagram, LinkedIn, páginas amarelas). Esse telefone achado na web é o que MAIS interessa — o telefone oficial da Receita Federal o usuário já vê em outra tela.
- Em "contatos_encontrados.telefone": só coloque um telefone que apareça LITERALMENTE nos trechos de busca acima E seja claramente da MESMA empresa (confira nome/endereço/cidade). Se houver qualquer dúvida, deixe null. NUNCA invente nem "complete" um número.
- Para site, instagram, linkedin e email: mesma regra — só se aparecer literalmente nos trechos. Se não encontrar, null.
- Para "pontos_fortes", "sinais_atencao", "abordagem_sugerida" e "perguntas_qualificacao": baseie-se nos dados da Receita Federal, sem inventar.
- Em "fontes", liste apenas os números [n] dos resultados de busca que você realmente usou para extrair algum contato.

Responda ESTRITAMENTE em JSON válido, sem nenhum texto antes ou depois, seguindo este formato exato:
{
  "resumo": "resumo de 2-3 frases sobre a empresa e o potencial comercial",
  "pontos_fortes": ["ponto 1", "ponto 2", "ponto 3"],
  "sinais_atencao": ["sinal 1", "sinal 2"],
  "abordagem_sugerida": "sugestão de abordagem comercial em 2-3 frases",
  "perguntas_qualificacao": ["pergunta 1", "pergunta 2", "pergunta 3"],
  "nivel_confianca": "alto, médio ou baixo",
  "contatos_encontrados": {
    "telefone": "string ou null",
    "site": "string ou null",
    "instagram": "string ou null",
    "linkedin": "string ou null",
    "email": "string ou null",
    "fontes": [1, 3]
  }
}`, e.Razao, e.NomeFantasia, e.Situacao, e.Porte, e.Natureza, e.Atividade, e.DataInicio, e.CapitalSocial,
		e.Municipio, e.UF, telOficialOuNaoInformado(telOficial), sociosTxt, blocoWeb)

	msg, err := s.groq.Chamar(ctx, []GroqMessage{{Role: "user", Content: prompt}}, ChamarOpcoes{
		Modelo: ModeloInsight, Temperatura: 0.3, JSONObjeto: true,
	})
	if err != nil {
		return nil, err
	}

	var parsed respostaGroqInsight
	if err := json.Unmarshal([]byte(msg.Content), &parsed); err != nil {
		return nil, fmt.Errorf("a IA retornou um JSON inválido: %w", err)
	}

	c := parsed.ContatosEncontrados
	fontes := make([]TavilyResultado, 0, len(c.Fontes))
	for _, n := range c.Fontes {
		if n >= 1 && n <= len(resultadosWeb) {
			fontes = append(fontes, resultadosWeb[n-1])
		}
	}
	telWeb := strFromPtr(c.Telefone)
	telefonePrincipal := telOficial
	if telefonePrincipal == "" {
		telefonePrincipal = telWeb
	}

	return &Insight{
		Resumo:                valorOuPadrao(parsed.Resumo, "Sem resumo disponível."),
		PontosFortes:          parsed.PontosFortes,
		SinaisAtencao:         parsed.SinaisAtencao,
		AbordagemSugerida:     parsed.AbordagemSugerida,
		PerguntasQualificacao: parsed.PerguntasQualificacao,
		NivelConfianca:        valorOuPadrao(parsed.NivelConfianca, "—"),
		BuscaWebRealizada:     len(resultadosWeb) > 0,
		Contatos: ContatosInsight{
			TelefoneOficial: telOficial,
			TelefoneWeb:     telWeb,
			Telefone:        telefonePrincipal,
			Site:            strFromPtr(c.Site),
			Instagram:       strFromPtr(c.Instagram),
			LinkedIn:        strFromPtr(c.LinkedIn),
			Email:           strFromPtr(c.Email),
			Fontes:          fontes,
		},
	}, nil
}

func semTraco(s string) string {
	if s == "" || s == "—" {
		return ""
	}
	return s
}

func telOficialOuNaoInformado(s string) string {
	if s == "" {
		return "não informado"
	}
	return s
}

func strFromPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func valorOuPadrao(s, padrao string) string {
	if strings.TrimSpace(s) == "" {
		return padrao
	}
	return s
}
