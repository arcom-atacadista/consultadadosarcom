package ia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const tavilyAPIURLPadrao = "https://api.tavily.com/search"

type TavilyResultado struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type tavilyRequisicao struct {
	APIKey        string `json:"api_key"`
	Query         string `json:"query"`
	SearchDepth   string `json:"search_depth"`
	IncludeAnswer bool   `json:"include_answer"`
	MaxResults    int    `json:"max_results"`
}

type tavilyResposta struct {
	Answer  string            `json:"answer"`
	Results []TavilyResultado `json:"results"`
}

type TavilyClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewTavilyClient recebe baseURL vazio pra usar o endpoint real da Tavily —
// só é sobrescrito (ex.: em teste, apontando pra um mock local).
func NewTavilyClient(apiKey, baseURL string) *TavilyClient {
	if baseURL == "" {
		baseURL = tavilyAPIURLPadrao
	}
	return &TavilyClient{apiKey: apiKey, baseURL: baseURL, httpClient: &http.Client{Timeout: 20 * time.Second}}
}

// Buscar faz uma busca real na web. incluirResumo pede o resumo (`answer`) da
// Tavily além dos resultados brutos.
func (c *TavilyClient) Buscar(ctx context.Context, query string, profundidade string, maxResultados int, incluirResumo bool) (string, []TavilyResultado, error) {
	if c.apiKey == "" {
		return "", nil, fmt.Errorf("TAVILY_API_KEY não configurada")
	}
	req := tavilyRequisicao{
		APIKey: c.apiKey, Query: query, SearchDepth: profundidade,
		IncludeAnswer: incluirResumo, MaxResults: maxResultados,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return "", nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return "", nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", nil, fmt.Errorf("chamar Tavily: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("Tavily: HTTP %d", resp.StatusCode)
	}

	var out tavilyResposta
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", nil, fmt.Errorf("decodificar resposta da Tavily: %w", err)
	}
	return out.Answer, out.Results, nil
}
