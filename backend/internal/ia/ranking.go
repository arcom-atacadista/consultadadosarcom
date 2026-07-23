package ia

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type EmpresaResumo struct {
	CNPJ          string `json:"cnpj"`
	Razao         string `json:"razao"`
	Situacao      string `json:"situacao"`
	Porte         string `json:"porte"`
	Atividade     string `json:"atividade"`
	Municipio     string `json:"municipio"`
	UF            string `json:"uf"`
	CapitalSocial string `json:"capitalSocial"`
}

type ItemRanking struct {
	CNPJ    string `json:"cnpj"`
	Razao   string `json:"razao"`
	Posicao int    `json:"posicao"`
	Motivo  string `json:"motivo"`
}

type respostaGroqRanking struct {
	Ranking []ItemRanking `json:"ranking"`
}

const maxEmpresasRanking = 60
const maxItensRanking = 15

type RankingService struct {
	groq *GroqClient
}

func NewRankingService(groq *GroqClient) *RankingService {
	return &RankingService{groq: groq}
}

// Gerar prioriza os leads mais promissores num único request — mais rápido
// e mais barato do que gerar insight individual de cada empresa.
func (s *RankingService) Gerar(ctx context.Context, empresas []EmpresaResumo) ([]ItemRanking, error) {
	if len(empresas) > maxEmpresasRanking {
		empresas = empresas[:maxEmpresasRanking]
	}

	linhas := make([]string, len(empresas))
	for i, e := range empresas {
		linhas[i] = fmt.Sprintf("%s|%s|%s|%s|%s|%s/%s|capital:%s",
			e.CNPJ, e.Razao, e.Situacao, e.Porte, e.Atividade, e.Municipio, e.UF, e.CapitalSocial)
	}

	prompt := fmt.Sprintf(`Você é um analista comercial. Abaixo está uma lista de empresas (formato: CNPJ|Razão Social|Situação|Porte|Atividade|Cidade/UF|Capital Social), uma por linha:

%s

Priorize as empresas mais promissoras como leads comerciais (considere: situação ATIVA é obrigatória para ser um bom lead, porte maior e capital social maior tendem a ser melhores, atividade compatível com potencial de compra recorrente). Responda APENAS com um JSON válido, sem markdown, no formato:
{"ranking":[{"cnpj":"...","razao":"...","posicao":1,"motivo":"frase curta explicando por que é um bom lead"}]}
Inclua no máximo os %d melhores, ordenados do melhor para o pior.`, strings.Join(linhas, "\n"), maxItensRanking)

	msg, err := s.groq.Chamar(ctx, []GroqMessage{
		{Role: "system", Content: "Você responde sempre em português do Brasil e apenas com JSON válido, sem texto adicional e sem markdown."},
		{Role: "user", Content: prompt},
	}, ChamarOpcoes{Modelo: ModeloInsight, Temperatura: 0.3, JSONObjeto: true})
	if err != nil {
		return nil, err
	}

	var parsed respostaGroqRanking
	if err := json.Unmarshal([]byte(msg.Content), &parsed); err != nil {
		return nil, fmt.Errorf("a IA retornou um JSON inválido: %w", err)
	}
	return parsed.Ranking, nil
}
