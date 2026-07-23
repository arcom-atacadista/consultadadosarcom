package ia

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const maxIteracoesTools = 5 // trava contra loop infinito de tool-calling

var toolsChat = []GroqTool{
	{
		Type: "function",
		Function: struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Parameters  any    `json:"parameters"`
		}{
			Name:        "buscar_web",
			Description: "Busca informações ATUAIS na internet (telefone, site, Instagram, endereço, notícias, avaliações, se a empresa funciona, preços, dados recentes). Use quando a resposta depender de informação externa/atual.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{"type": "string", "description": "O termo de busca, o mais específico possível"},
				},
				"required": []string{"query"},
			},
		},
	},
	{
		Type: "function",
		Function: struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Parameters  any    `json:"parameters"`
		}{
			Name:        "estatisticas_do_site",
			Description: "Retorna os NÚMEROS/RELATÓRIOS do próprio site CDA em tempo real: contas criadas/pendentes, quantidade de consultas e prospecções. Use SEMPRE que o usuário perguntar sobre o uso do site, quantas consultas/prospecções, etc.",
			Parameters:  map[string]any{"type": "object", "properties": map[string]any{}, "required": []string{}},
		},
	},
}

type MensagemChat struct {
	Role  string `json:"role"`
	Texto string `json:"texto"`
}

// EstatisticasFn busca os números reais do site — injetada de fora (ia não
// conhece usuarios/cnpj/prospeccao diretamente) e já filtrada por permissão
// pelo chamador (só admin recebe números de verdade).
type EstatisticasFn func(ctx context.Context) (string, error)

type ChatService struct {
	groq         *GroqClient
	tavily       *TavilyClient
	estatisticas EstatisticasFn
}

func NewChatService(groq *GroqClient, tavily *TavilyClient, estatisticas EstatisticasFn) *ChatService {
	return &ChatService{groq: groq, tavily: tavily, estatisticas: estatisticas}
}

func montarSystemPrompt(empresasContexto []EmpresaResumo) string {
	hoje := time.Now().Format("02/01/2006")
	contexto := ""
	if len(empresasContexto) > 0 {
		linhas := make([]string, 0, len(empresasContexto))
		for _, e := range empresasContexto {
			linhas = append(linhas, fmt.Sprintf("- %s (%s) — %s, %s, %s/%s, atividade: %s",
				e.Razao, e.CNPJ, e.Situacao, e.Porte, e.Municipio, e.UF, e.Atividade))
		}
		contexto = "\n\nO usuário já consultou estas empresas nesta sessão (dados públicos da Receita Federal):\n" + strings.Join(linhas, "\n")
	}

	return fmt.Sprintf(`Você é o assistente de IA do CDA (Consulta Dados Arcom), ferramenta da Arcom para consulta de CNPJ e prospecção. Responda em português do Brasil, de forma clara, objetiva e útil. Hoje é %s.

VOCÊ TEM ACESSO À INTERNET pela ferramenta "buscar_web" — e deve USÁ-LA de verdade:
- Sempre que a pergunta envolver QUALQUER informação externa, atual ou que você não tenha 100%% de certeza — telefone, endereço, site, Instagram/redes sociais, horário de funcionamento, se uma loja/empresa existe ou ainda funciona, notícias, preços, nomes de pessoas/empresas, eventos — FAÇA a busca ANTES de responder. Na dúvida, busque.
- Você pode buscar VÁRIAS vezes e refinar: se o 1º resultado não bastar, tente outra query.
- Responda com base nos resultados e SEMPRE cite as fontes: liste os links (URLs) que embasaram a resposta no final.
- NUNCA diga que "não tem acesso à internet" ou que "não pode navegar" — você PODE. Se realmente não achar, diga que buscou e não encontrou.
- Só responda direto (sem buscar) dúvidas simples ou sobre os dados de empresas já consultados na ferramenta (listados abaixo).

VOCÊ TAMBÉM CONHECE OS RELATÓRIOS DO PRÓPRIO SITE pela ferramenta "estatisticas_do_site". Sempre que perguntarem sobre o uso do site (contas, consultas, prospecções), CHAME "estatisticas_do_site" e responda com os números reais. Não invente números.%s`, hoje, contexto)
}

// Responder roda o loop de tool-calling: a IA pode pedir várias buscas na web
// (ou as estatísticas do site) e refinar até ter o que precisa, no máximo
// maxIteracoesTools vezes.
func (s *ChatService) Responder(ctx context.Context, mensagemUsuario string, historico []MensagemChat, empresasContexto []EmpresaResumo, isAdmin bool) (string, error) {
	messages := []GroqMessage{{Role: "system", Content: montarSystemPrompt(empresasContexto)}}
	inicio := 0
	if len(historico) > 12 {
		inicio = len(historico) - 12
	}
	for _, m := range historico[inicio:] {
		messages = append(messages, GroqMessage{Role: m.Role, Content: m.Texto})
	}
	messages = append(messages, GroqMessage{Role: "user", Content: mensagemUsuario})

	for iter := 0; iter < maxIteracoesTools; iter++ {
		msg, err := s.groq.Chamar(ctx, messages, ChamarOpcoes{
			Modelo: ModeloChat, Temperatura: 0.5, Tools: toolsChat, ToolChoiceAuto: true,
		})
		if err != nil {
			return "", err
		}
		if len(msg.ToolCalls) == 0 {
			if msg.Content == "" {
				return "Desculpe, não consegui gerar uma resposta.", nil
			}
			return msg.Content, nil
		}

		messages = append(messages, *msg)
		for _, tc := range msg.ToolCalls {
			resultado := s.executarTool(ctx, tc, mensagemUsuario, isAdmin)
			messages = append(messages, GroqMessage{
				Role: "tool", ToolCallID: tc.ID, Name: tc.Function.Name, Content: resultado,
			})
		}
	}

	// esgotou as iterações — pede uma resposta final sem mais ferramentas
	msg, err := s.groq.Chamar(ctx, messages, ChamarOpcoes{Modelo: ModeloChat, Temperatura: 0.5})
	if err != nil {
		return "", err
	}
	if msg.Content == "" {
		return "Desculpe, não consegui gerar uma resposta.", nil
	}
	return msg.Content, nil
}

func (s *ChatService) executarTool(ctx context.Context, tc GroqToolCall, mensagemOriginal string, isAdmin bool) string {
	switch tc.Function.Name {
	case "estatisticas_do_site":
		if !isAdmin {
			return "Estatísticas do site são visíveis só para administradores — explique isso ao usuário."
		}
		if s.estatisticas == nil {
			return "Estatísticas indisponíveis no momento."
		}
		texto, err := s.estatisticas(ctx)
		if err != nil {
			return "Erro ao buscar estatísticas: " + err.Error()
		}
		return texto
	default: // buscar_web
		var args struct {
			Query string `json:"query"`
		}
		json.Unmarshal([]byte(tc.Function.Arguments), &args)
		query := args.Query
		if query == "" {
			query = mensagemOriginal
		}
		return s.buscarWeb(ctx, query)
	}
}

func (s *ChatService) buscarWeb(ctx context.Context, query string) string {
	answer, resultados, err := s.tavily.Buscar(ctx, query, "basic", 6, true)
	if err != nil {
		return "Erro na busca web: " + err.Error()
	}
	var out strings.Builder
	if answer != "" {
		out.WriteString("Resumo da web: " + answer + "\n\n")
	}
	for i, r := range resultados {
		if i >= 6 {
			break
		}
		content := r.Content
		if len(content) > 300 {
			content = content[:300]
		}
		fmt.Fprintf(&out, "[%d] %s\n%s\n%s\n\n", i+1, r.Title, r.URL, content)
	}
	if out.Len() == 0 {
		return "Nenhum resultado encontrado na web."
	}
	return strings.TrimSpace(out.String())
}
