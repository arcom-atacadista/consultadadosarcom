// Package ia fala com Groq (chat completions) e Tavily (busca web) — as
// chaves nunca saem do backend (docs/migracao/01 §3.5). Substitui o
// Cloudflare Worker que o app antigo usava como proxy.
package ia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const groqAPIURLPadrao = "https://api.groq.com/openai/v1/chat/completions"

// Modelos usados — 70B pra insight/ranking (mais inteligente), 8B pro chat
// (mais rápido, limite diário maior).
const (
	ModeloInsight = "llama-3.3-70b-versatile"
	ModeloChat    = "llama-3.1-8b-instant"
)

type GroqMessage struct {
	Role       string         `json:"role"`
	Content    string         `json:"content,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	Name       string         `json:"name,omitempty"`
	ToolCalls  []GroqToolCall `json:"tool_calls,omitempty"`
}

type GroqToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type GroqTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Parameters  any    `json:"parameters"`
	} `json:"function"`
}

type groqRequisicao struct {
	Model          string        `json:"model"`
	Messages       []GroqMessage `json:"messages"`
	Temperature    float64       `json:"temperature"`
	ResponseFormat any           `json:"response_format,omitempty"`
	Tools          []GroqTool    `json:"tools,omitempty"`
	ToolChoice     string        `json:"tool_choice,omitempty"`
}

type groqResposta struct {
	Choices []struct {
		Message GroqMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type GroqClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewGroqClient recebe baseURL vazio pra usar o endpoint real da Groq —
// só é sobrescrito (ex.: em teste, apontando pra um mock local).
func NewGroqClient(apiKey, baseURL string) *GroqClient {
	if baseURL == "" {
		baseURL = groqAPIURLPadrao
	}
	return &GroqClient{apiKey: apiKey, baseURL: baseURL, httpClient: &http.Client{Timeout: 45 * time.Second}}
}

type ChamarOpcoes struct {
	Modelo         string
	Temperatura    float64
	JSONObjeto     bool
	Tools          []GroqTool
	ToolChoiceAuto bool
}

// Chamar faz uma chamada de chat completions ao Groq e devolve a mensagem de
// resposta (pode conter tool_calls, se `opts.Tools` foi passado).
func (c *GroqClient) Chamar(ctx context.Context, messages []GroqMessage, opts ChamarOpcoes) (*GroqMessage, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY não configurada")
	}
	req := groqRequisicao{
		Model:       opts.Modelo,
		Messages:    messages,
		Temperature: opts.Temperatura,
		Tools:       opts.Tools,
	}
	if opts.JSONObjeto {
		req.ResponseFormat = map[string]string{"type": "json_object"}
	}
	if opts.ToolChoiceAuto {
		req.ToolChoice = "auto"
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("chamar Groq: %w", err)
	}
	defer resp.Body.Close()

	var out groqResposta
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decodificar resposta do Groq: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("HTTP %d", resp.StatusCode)
		if out.Error != nil {
			msg = out.Error.Message
		}
		return nil, fmt.Errorf("Groq: %s", msg)
	}
	if len(out.Choices) == 0 {
		return nil, fmt.Errorf("Groq não retornou nenhuma escolha")
	}
	return &out.Choices[0].Message, nil
}
