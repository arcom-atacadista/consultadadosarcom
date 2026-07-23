// Package enriquecimento fala com a Trace360 (dossiê de enriquecimento por
// CNPJ) — a x-api-key nunca sai do backend (docs/migracao/01 §3.3).
package enriquecimento

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ItemEnviado struct {
	CNPJ      string `json:"cnpj"`
	ClienteID string `json:"cliente_id"`
}

type respostaEnvio struct {
	Resultados []ItemEnviado `json:"resultados"`
	Mensagem   string        `json:"mensagem"`
}

type StatusCliente struct {
	Status      string `json:"status"`
	CNPJ        string `json:"cnpj"`
	RazaoSocial string `json:"razao_social"`
	CriadoEm    string `json:"criado_em"`
}

type EventoProgresso struct {
	Etapa  string `json:"etapa"`
	Status string `json:"status"`
}

type Progresso struct {
	Status     string            `json:"status"`
	EtapaAtual string            `json:"etapa_atual"`
	CNPJ       string            `json:"cnpj"`
	Eventos    []EventoProgresso `json:"eventos"`
}

type Trace360Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewTrace360Client(baseURL, apiKey string) *Trace360Client {
	return &Trace360Client{baseURL: baseURL, apiKey: apiKey, httpClient: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Trace360Client) fazer(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.httpClient.Do(req)
}

// EnviarLote enfileira os CNPJs pra enriquecimento (até 500 por chamada).
func (c *Trace360Client) EnviarLote(ctx context.Context, cnpjs []string) ([]ItemEnviado, error) {
	payload := make([]map[string]string, len(cnpjs))
	for i, c := range cnpjs {
		payload[i] = map[string]string{"cnpj": c}
	}
	resp, err := c.fazer(ctx, http.MethodPost, "/clientes", payload)
	if err != nil {
		return nil, fmt.Errorf("enviar CNPJs pra Trace360: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		corpo, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		return nil, fmt.Errorf("Trace360: HTTP %d %s", resp.StatusCode, corpo)
	}
	var out respostaEnvio
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decodificar resposta da Trace360: %w", err)
	}
	return out.Resultados, nil
}

// Status busca o status atual de um cliente enriquecido.
func (c *Trace360Client) Status(ctx context.Context, clienteID string) (*StatusCliente, error) {
	resp, err := c.fazer(ctx, http.MethodGet, "/clientes/"+clienteID, nil)
	if err != nil {
		return nil, fmt.Errorf("buscar status na Trace360: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Trace360: HTTP %d", resp.StatusCode)
	}
	var out StatusCliente
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decodificar status da Trace360: %w", err)
	}
	return &out, nil
}

// Progresso busca o fluxo de etapas (busca, desambiguação, análise, PDF...).
func (c *Trace360Client) Progresso(ctx context.Context, clienteID string) (*Progresso, error) {
	resp, err := c.fazer(ctx, http.MethodGet, "/clientes/"+clienteID+"/progresso", nil)
	if err != nil {
		return nil, fmt.Errorf("buscar progresso na Trace360: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Trace360: HTTP %d", resp.StatusCode)
	}
	var out Progresso
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decodificar progresso da Trace360: %w", err)
	}
	return &out, nil
}

// Dossie baixa o PDF do dossiê (bytes + content-type).
func (c *Trace360Client) Dossie(ctx context.Context, clienteID string) ([]byte, string, error) {
	resp, err := c.fazer(ctx, http.MethodGet, "/clientes/"+clienteID+"/dossie", nil)
	if err != nil {
		return nil, "", fmt.Errorf("baixar dossiê da Trace360: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("Trace360: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return body, resp.Header.Get("Content-Type"), nil
}

// Resultado busca os dados brutos (JSON) do enriquecimento concluído.
func (c *Trace360Client) Resultado(ctx context.Context, clienteID string) (json.RawMessage, error) {
	resp, err := c.fazer(ctx, http.MethodGet, "/clientes/"+clienteID+"/resultado", nil)
	if err != nil {
		return nil, fmt.Errorf("buscar resultado na Trace360: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Trace360: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// Reprocessar manda tentar de novo um enriquecimento que falhou.
func (c *Trace360Client) Reprocessar(ctx context.Context, clienteID string) error {
	resp, err := c.fazer(ctx, http.MethodPost, "/clientes/"+clienteID+"/reprocessar", map[string]any{})
	if err != nil {
		return fmt.Errorf("reprocessar na Trace360: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Trace360: HTTP %d", resp.StatusCode)
	}
	return nil
}
