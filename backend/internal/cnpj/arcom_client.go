package cnpj

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	arcomLoteTamanhoMax = 1000
	arcomTokenFolga     = 89 * 24 * time.Hour // token vale ~90 dias, guardamos com 1 dia de folga
)

// comDescricao é o formato {codigo, descricao} que a API Arcom usa pra vários
// campos (situação cadastral, natureza jurídica, porte, CNAE, município...).
type comDescricao struct {
	Codigo    string `json:"codigo"`
	Descricao string `json:"descricao"`
}

type socioBruto struct {
	NomeSocio            string       `json:"nome_socio"`
	CnpjCpfSocio         string       `json:"cnpj_cpf_socio"`
	QualificacaoSocio    comDescricao `json:"qualificacao_socio"`
	FaixaEtaria          comDescricao `json:"faixa_etaria"`
	DataEntradaSociedade string       `json:"data_entrada_sociedade"`
}

type empresaBruta struct {
	RazaoSocial      string       `json:"razao_social"`
	PorteEmpresa     comDescricao `json:"porte_empresa"`
	NaturezaJuridica comDescricao `json:"natureza_juridica"`
	CapitalSocial    *float64     `json:"capital_social"`
}

type simplesBruto struct {
	OpcaoSimples any `json:"opcao_simples"`
	OpcaoMei     any `json:"opcao_mei"`
}

type dadosBrutosArcom struct {
	SituacaoCadastral       comDescricao `json:"situacao_cadastral"`
	DataSituacaoCadastral   string       `json:"data_situacao_cadastral"`
	MotivoSituacaoCadastral comDescricao `json:"motivo_situacao_cadastral"`
	Empresa                 empresaBruta `json:"empresa"`
	CnaeFiscalPrincipal     comDescricao `json:"cnae_fiscal_principal"`
	MatrizFilial            comDescricao `json:"matriz_filial"`
	UF                      string       `json:"uf"`
	NomeFantasia            string       `json:"nome_fantasia"`
	DataInicioAtividade     string       `json:"data_inicio_atividade"`
	TipoLogradouro          string       `json:"tipo_logradouro"`
	Logradouro              string       `json:"logradouro"`
	Numero                  string       `json:"numero"`
	Complemento             string       `json:"complemento"`
	Bairro                  string       `json:"bairro"`
	Municipio               comDescricao `json:"municipio"`
	CEP                     string       `json:"cep"`
	DDD1                    string       `json:"ddd_1"`
	Telefone1               string       `json:"telefone_1"`
	CorreioEletronico       string       `json:"correio_eletronico"`
	Socios                  []socioBruto `json:"socios"`
	Simples                 simplesBruto `json:"simples"`
	Latitude                *float64     `json:"latitude"`
	Longitude               *float64     `json:"longitude"`
}

type itemLoteArcom struct {
	CNPJ           string           `json:"cnpj"`
	Found          bool             `json:"found"`
	Error          string           `json:"error"`
	Data           dadosBrutosArcom `json:"data"`
	JaCliente      bool             `json:"ja_cliente"`
	IDOcorCadastro any              `json:"id_ocor_cadastro"`
}

type respostaBatchArcom struct {
	Results []itemLoteArcom `json:"results"`
}

// ArcomClient fala com a Consulta CNPJ Arcom (docs/migracao/01 §3.1). A
// api_key nunca sai do backend — o navegador não vê nada disso.
type ArcomClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client

	mu       sync.Mutex
	token    string
	tokenExp time.Time
}

func NewArcomClient(baseURL, apiKey string) *ArcomClient {
	return &ArcomClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *ArcomClient) garantirToken(ctx context.Context, forcar bool) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !forcar && c.token != "" && time.Now().Before(c.tokenExp) {
		return c.token, nil
	}

	body, _ := json.Marshal(map[string]string{"api_key": c.apiKey})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/auth/token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("autenticar na Consulta CNPJ Arcom: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("autenticar na Consulta CNPJ Arcom: HTTP %d", resp.StatusCode)
	}

	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decodificar token da Arcom: %w", err)
	}

	c.token = out.Token
	c.tokenExp = time.Now().Add(arcomTokenFolga)
	return c.token, nil
}

// ConsultarLote consulta um lote de CNPJs (chunk automático de até 1000),
// reautenticando uma vez se o token expirou (401).
func (c *ArcomClient) ConsultarLote(ctx context.Context, cnpjs []string) ([]itemLoteArcom, error) {
	var resultados []itemLoteArcom
	for _, chunk := range dividirEmChunks(cnpjs, arcomLoteTamanhoMax) {
		itens, err := c.consultarChunk(ctx, chunk)
		if err != nil {
			return nil, err
		}
		resultados = append(resultados, itens...)
	}
	return resultados, nil
}

func (c *ArcomClient) consultarChunk(ctx context.Context, cnpjs []string) ([]itemLoteArcom, error) {
	token, err := c.garantirToken(ctx, false)
	if err != nil {
		return nil, err
	}

	fazer := func(tok string) (*http.Response, error) {
		payload, _ := json.Marshal(map[string]any{"cnpjs": cnpjs, "cliente": true})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/cnpj/batch", bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("Authorization", "Bearer "+tok)
		return c.httpClient.Do(req)
	}

	resp, err := fazer(token)
	if err != nil {
		return nil, fmt.Errorf("consultar CNPJ na Arcom: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		token, err = c.garantirToken(ctx, true)
		if err != nil {
			return nil, err
		}
		resp, err = fazer(token)
		if err != nil {
			return nil, fmt.Errorf("consultar CNPJ na Arcom (retry): %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		corpo, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		return nil, fmt.Errorf("consultar CNPJ na Arcom: HTTP %d %s", resp.StatusCode, corpo)
	}

	var out respostaBatchArcom
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decodificar resposta da Arcom: %w", err)
	}
	return out.Results, nil
}

func dividirEmChunks(cnpjs []string, tamanho int) [][]string {
	var chunks [][]string
	for i := 0; i < len(cnpjs); i += tamanho {
		fim := i + tamanho
		if fim > len(cnpjs) {
			fim = len(cnpjs)
		}
		chunks = append(chunks, cnpjs[i:fim])
	}
	return chunks
}

// mapArcomItem normaliza o formato aninhado da Arcom pro DTO único da API
// (mesma lógica do mapArcomBatchRecord do app antigo).
func mapArcomItem(item itemLoteArcom) Empresa {
	if !item.Found {
		msg := item.Error
		if msg == "" {
			msg = "CNPJ não encontrado"
		}
		return Empresa{CNPJ: item.CNPJ, Encontrado: false, Erro: msg, API: "Consulta CNPJ Arcom"}
	}

	d := item.Data
	nomeMunicipio := d.Municipio.Descricao

	enderecoInicio := juntarNaoVazios(", ", juntarNaoVazios(" ", d.TipoLogradouro, d.Logradouro), d.Numero)
	complemento := ""
	if d.Complemento != "" {
		complemento = " (" + d.Complemento + ")"
	}
	enderecoFim := juntarNaoVazios(", ", d.Bairro, nomeMunicipio, d.UF)
	endereco := "—"
	if enderecoInicio != "" || enderecoFim != "" {
		endereco = enderecoInicio + complemento
		if enderecoFim != "" {
			endereco += " - " + enderecoFim
		}
	}

	situacao := d.SituacaoCadastral.Descricao
	if situacao == "" {
		situacao = "INDEFINIDA"
	}

	telefone := "—"
	if d.Telefone1 != "" {
		telefone = d.DDD1 + d.Telefone1
	}

	socios := make([]Socio, 0, len(d.Socios))
	for _, s := range d.Socios {
		socios = append(socios, Socio{
			NomeSocio:            s.NomeSocio,
			CPF:                  s.CnpjCpfSocio,
			QualificacaoSocio:    s.QualificacaoSocio.Descricao,
			FaixaEtaria:          s.FaixaEtaria.Descricao,
			DataEntradaSociedade: s.DataEntradaSociedade,
		})
	}

	return Empresa{
		CNPJ:                    item.CNPJ,
		Encontrado:              true,
		Situacao:                strUpper(situacao),
		DataSituacaoCadastral:   comOuTraco(d.DataSituacaoCadastral),
		MotivoSituacaoCadastral: comOuTraco(d.MotivoSituacaoCadastral.Descricao),
		Razao:                   comOuTraco(d.Empresa.RazaoSocial),
		NomeFantasia:            comOuTraco(d.NomeFantasia),
		Porte:                   comOuTraco(d.Empresa.PorteEmpresa.Descricao),
		ClienteArcom:            simNaoOuTraco(item.JaCliente),
		Natureza:                comOuTraco(d.Empresa.NaturezaJuridica.Descricao),
		Atividade:               comOuTraco(d.CnaeFiscalPrincipal.Descricao),
		MatrizFilial:            comOuTraco(d.MatrizFilial.Descricao),
		UF:                      d.UF,
		Municipio:               nomeMunicipio,
		DataInicio:              comOuTraco(d.DataInicioAtividade),
		CapitalSocial:           capitalOuTraco(d.Empresa.CapitalSocial),
		Endereco:                endereco,
		CEP:                     comOuTraco(d.CEP),
		Telefone:                telefone,
		Email:                   comOuTraco(d.CorreioEletronico),
		Socios:                  socios,
		Simples:                 simNaoOuTraco(d.Simples.OpcaoSimples),
		MEI:                     simNaoOuTraco(d.Simples.OpcaoMei),
		Latitude:                d.Latitude,
		Longitude:               d.Longitude,
		API:                     "Consulta CNPJ Arcom",
	}
}
