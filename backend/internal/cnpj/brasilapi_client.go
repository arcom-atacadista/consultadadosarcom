package cnpj

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const brasilAPIBaseURL = "https://brasilapi.com.br/api/cnpj/v1"

type cnaeSecundarioBrasilAPI struct {
	Codigo    int    `json:"codigo"`
	Descricao string `json:"descricao"`
}

type socioBrasilAPI struct {
	NomeSocio            string `json:"nome_socio"`
	CnpjCpfDoSocio       string `json:"cnpj_cpf_do_socio"`
	QualificacaoSocio    string `json:"qualificacao_socio"`
	FaixaEtaria          string `json:"faixa_etaria"`
	DataEntradaSociedade string `json:"data_entrada_sociedade"`
}

type respostaBrasilAPI struct {
	UF                                 string                    `json:"uf"`
	CEP                                string                    `json:"cep"`
	QSA                                []socioBrasilAPI          `json:"qsa"`
	CNAEsSecundarios                   []cnaeSecundarioBrasilAPI `json:"cnaes_secundarios"`
	RazaoSocial                        string                    `json:"razao_social"`
	NomeFantasia                       string                    `json:"nome_fantasia"`
	Porte                              string                    `json:"porte"`
	NaturezaJuridica                   string                    `json:"natureza_juridica"`
	CnaeFiscalDescricao                string                    `json:"cnae_fiscal_descricao"`
	DescricaoIdentificadorMatrizFilial string                    `json:"descricao_identificador_matriz_filial"`
	DescricaoSituacaoCadastral         string                    `json:"descricao_situacao_cadastral"`
	DataSituacaoCadastral              string                    `json:"data_situacao_cadastral"`
	DescricaoMotivoSituacaoCadastral   string                    `json:"descricao_motivo_situacao_cadastral"`
	DataInicioAtividade                string                    `json:"data_inicio_atividade"`
	CapitalSocial                      *float64                  `json:"capital_social"`
	DescricaoTipoDeLogradouro          string                    `json:"descricao_tipo_de_logradouro"`
	Logradouro                         string                    `json:"logradouro"`
	Numero                             string                    `json:"numero"`
	Complemento                        string                    `json:"complemento"`
	Bairro                             string                    `json:"bairro"`
	Municipio                          string                    `json:"municipio"`
	DDDTelefone1                       string                    `json:"ddd_telefone_1"`
	Email                              string                    `json:"email"`
	OpcaoPeloSimples                   *bool                     `json:"opcao_pelo_simples"`
	OpcaoPeloMei                       *bool                     `json:"opcao_pelo_mei"`
}

type BrasilAPIClient struct {
	httpClient *http.Client
}

func NewBrasilAPIClient() *BrasilAPIClient {
	return &BrasilAPIClient{httpClient: &http.Client{Timeout: 15 * time.Second}}
}

// Consultar busca 1 CNPJ por vez — é o limite da BrasilAPI (pública, sem chave).
func (c *BrasilAPIClient) Consultar(ctx context.Context, cnpjLimpo string) (Empresa, error) {
	url := fmt.Sprintf("%s/%s", brasilAPIBaseURL, cnpjLimpo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Empresa{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Empresa{CNPJ: cnpjLimpo, Encontrado: false, Erro: "falha ao contatar a Brasil API", API: "Brasil API"}, nil
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return Empresa{CNPJ: cnpjLimpo, Encontrado: false, Erro: "CNPJ não encontrado", API: "Brasil API"}, nil
	case http.StatusTooManyRequests:
		return Empresa{CNPJ: cnpjLimpo, Encontrado: false, Erro: "Brasil API: muitas requisições, aguarde um instante", API: "Brasil API"}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return Empresa{CNPJ: cnpjLimpo, Encontrado: false, Erro: fmt.Sprintf("Brasil API: erro HTTP %d", resp.StatusCode), API: "Brasil API"}, nil
	}

	var d respostaBrasilAPI
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return Empresa{}, fmt.Errorf("decodificar resposta da Brasil API: %w", err)
	}
	return mapBrasilAPIRecord(cnpjLimpo, d), nil
}

func mapBrasilAPIRecord(cnpjLimpo string, d respostaBrasilAPI) Empresa {
	enderecoInicio := juntarNaoVazios(", ", juntarNaoVazios(" ", d.DescricaoTipoDeLogradouro, d.Logradouro), d.Numero)
	complemento := ""
	if d.Complemento != "" {
		complemento = " (" + d.Complemento + ")"
	}
	enderecoFim := juntarNaoVazios(", ", d.Bairro, d.Municipio, d.UF)
	endereco := "—"
	if enderecoInicio != "" || enderecoFim != "" {
		endereco = enderecoInicio + complemento
		if enderecoFim != "" {
			endereco += " - " + enderecoFim
		}
	}

	socios := make([]Socio, 0, len(d.QSA))
	for _, s := range d.QSA {
		socios = append(socios, Socio{
			NomeSocio:            s.NomeSocio,
			CPF:                  s.CnpjCpfDoSocio,
			QualificacaoSocio:    s.QualificacaoSocio,
			FaixaEtaria:          s.FaixaEtaria,
			DataEntradaSociedade: s.DataEntradaSociedade,
		})
	}

	situacao := d.DescricaoSituacaoCadastral
	if situacao == "" {
		situacao = "INDEFINIDA"
	}

	return Empresa{
		CNPJ:                    cnpjLimpo,
		Encontrado:              true,
		Situacao:                strUpper(situacao),
		DataSituacaoCadastral:   comOuTraco(d.DataSituacaoCadastral),
		MotivoSituacaoCadastral: comOuTraco(d.DescricaoMotivoSituacaoCadastral),
		Razao:                   comOuTraco(d.RazaoSocial),
		NomeFantasia:            comOuTraco(d.NomeFantasia),
		Porte:                   comOuTraco(d.Porte),
		ClienteArcom:            "—", // Brasil API não sabe se é cliente Arcom
		Natureza:                comOuTraco(d.NaturezaJuridica),
		Atividade:               comOuTraco(d.CnaeFiscalDescricao),
		MatrizFilial:            comOuTraco(d.DescricaoIdentificadorMatrizFilial),
		UF:                      d.UF,
		Municipio:               d.Municipio,
		DataInicio:              comOuTraco(d.DataInicioAtividade),
		CapitalSocial:           capitalOuTraco(d.CapitalSocial),
		Endereco:                endereco,
		CEP:                     comOuTraco(d.CEP),
		Telefone:                comOuTraco(d.DDDTelefone1),
		Email:                   comOuTraco(d.Email),
		Socios:                  socios,
		Simples:                 simNaoOuTraco(ptrBoolToAny(d.OpcaoPeloSimples)),
		MEI:                     simNaoOuTraco(ptrBoolToAny(d.OpcaoPeloMei)),
		Latitude:                nil,
		Longitude:               nil,
		API:                     "Brasil API",
	}
}

func ptrBoolToAny(b *bool) any {
	if b == nil {
		return nil
	}
	return *b
}
