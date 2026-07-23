package cnpj

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	arcomPageSize   = 1000 // tamanho de página aceito pela API por chamada
	arcomMaxPaginas = 200  // trava de segurança (200 x 1000 = até 200.000 registros)
)

// EstabelecimentoBruto é o formato (mais "achatado" que /v1/cnpj/batch, sem
// o wrapper "empresa") que a Arcom devolve em /v1/estabelecimentos.
type EstabelecimentoBruto struct {
	CNPJ                string       `json:"cnpj"`
	RazaoSocial         string       `json:"razao_social"`
	NomeFantasia        string       `json:"nome_fantasia"`
	CnaeFiscalPrincipal comDescricao `json:"cnae_fiscal_principal"`
	Bairro              string       `json:"bairro"`
	Logradouro          string       `json:"logradouro"`
	Numero              string       `json:"numero"`
	CEP                 string       `json:"cep"`
	DDD1                string       `json:"ddd_1"`
	Telefone1           string       `json:"telefone_1"`
	CorreioEletronico   string       `json:"correio_eletronico"`
	PorteEmpresa        comDescricao `json:"porte_empresa"`
	DataInicioAtividade string       `json:"data_inicio_atividade"`
	Latitude            *float64     `json:"latitude"`
	Longitude           *float64     `json:"longitude"`
}

type respostaEstabelecimentos struct {
	Data []EstabelecimentoBruto `json:"data"`
}

type EstabelecimentosParams struct {
	UF              string
	MunicipioCodigo string
	CNAE            string
}

// EstabelecimentosPagina busca uma página de estabelecimentos ativos e que
// ainda não são clientes Arcom (excluir_clientes=true, fail-closed em 503).
func (c *ArcomClient) EstabelecimentosPagina(ctx context.Context, p EstabelecimentosParams, offset int) ([]EstabelecimentoBruto, error) {
	token, err := c.garantirToken(ctx, false)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	if p.UF != "" {
		q.Set("uf", p.UF)
	}
	if p.MunicipioCodigo != "" {
		q.Set("municipio", p.MunicipioCodigo)
	}
	if p.CNAE != "" {
		q.Set("cnae_fiscal_principal", p.CNAE)
	}
	q.Set("situacao_cadastral", "02") // ATIVA
	q.Set("excluir_clientes", "true")
	q.Set("limit", fmt.Sprintf("%d", arcomPageSize))
	q.Set("offset", fmt.Sprintf("%d", offset))

	fazer := func(tok string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/estabelecimentos?"+q.Encode(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+tok)
		return c.httpClient.Do(req)
	}

	resp, err := fazer(token)
	if err != nil {
		return nil, fmt.Errorf("buscar estabelecimentos: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		token, err = c.garantirToken(ctx, true)
		if err != nil {
			return nil, err
		}
		resp, err = fazer(token)
		if err != nil {
			return nil, fmt.Errorf("buscar estabelecimentos (retry): %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusServiceUnavailable {
		return nil, fmt.Errorf("serviço de clientes indisponível no momento — tente novamente em instantes")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("buscar estabelecimentos: HTTP %d", resp.StatusCode)
	}

	var out respostaEstabelecimentos
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decodificar resposta de estabelecimentos: %w", err)
	}
	return out.Data, nil
}

// EstabelecimentosTodos busca todas as páginas disponíveis (sem limite de
// quantidade), parando quando uma página vier incompleta ou vazia.
func (c *ArcomClient) EstabelecimentosTodos(ctx context.Context, p EstabelecimentosParams) ([]EstabelecimentoBruto, error) {
	var todos []EstabelecimentoBruto
	offset := 0
	for pagina := 0; pagina < arcomMaxPaginas; pagina++ {
		parcial, err := c.EstabelecimentosPagina(ctx, p, offset)
		if err != nil {
			return nil, err
		}
		todos = append(todos, parcial...)
		if len(parcial) < arcomPageSize {
			break
		}
		offset += arcomPageSize
	}
	return todos, nil
}
