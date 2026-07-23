package cnpj

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// MunicipioBruto é o formato que a Arcom devolve em /v1/municipios.
type MunicipioBruto struct {
	Codigo    string `json:"codigo"`
	Descricao string `json:"descricao"`
	UF        string `json:"uf"`
}

type respostaMunicipios struct {
	Data []MunicipioBruto `json:"data"`
}

// Municipios resolve um nome de cidade em candidatos de município da Receita
// Federal (docs/migracao/01 §3.1). A desambiguação por UF fica a cargo de
// quem chama (internal/prospeccao) — aqui só repassamos o que a API devolve.
func (c *ArcomClient) Municipios(ctx context.Context, q string) ([]MunicipioBruto, error) {
	token, err := c.garantirToken(ctx, false)
	if err != nil {
		return nil, err
	}

	fazer := func(tok string) (*http.Response, error) {
		u := c.baseURL + "/v1/municipios?" + url.Values{"q": {q}}.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+tok)
		return c.httpClient.Do(req)
	}

	resp, err := fazer(token)
	if err != nil {
		return nil, fmt.Errorf("resolver município: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		token, err = c.garantirToken(ctx, true)
		if err != nil {
			return nil, err
		}
		resp, err = fazer(token)
		if err != nil {
			return nil, fmt.Errorf("resolver município (retry): %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("resolver município: HTTP %d", resp.StatusCode)
	}
	var out respostaMunicipios
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decodificar resposta de municípios: %w", err)
	}
	return out.Data, nil
}
