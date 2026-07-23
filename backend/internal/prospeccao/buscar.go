package prospeccao

import (
	"context"
	"fmt"
	"strings"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/cnpj"
)

// radarMesesNova: "nova" = aberta nos últimos 12 meses (mesmo limiar do app antigo).
const radarMesesNova = 12

// cnaesMisto é a cesta de ramos "Misto" (comércio de bairro em geral) do app antigo.
var cnaesMisto = []string{
	"4711302", "4771701", "5611201", "4520001", "4789004",
	"4761003", "4731800", "4721102", "8630503", "6821801",
}

// ExpandirCNAEs troca o valor especial "MISTO" pela cesta fixa de ramos e
// deduplica o resultado.
func ExpandirCNAEs(cnaes []string) []string {
	vistos := map[string]bool{}
	var out []string
	for _, c := range cnaes {
		alvo := []string{c}
		if c == "MISTO" {
			alvo = cnaesMisto
		}
		for _, a := range alvo {
			if !vistos[a] {
				vistos[a] = true
				out = append(out, a)
			}
		}
	}
	return out
}

// Buscador orquestra a busca de prospecção contra a API Arcom.
type Buscador struct {
	arcom *cnpj.ArcomClient
}

func NewBuscador(arcom *cnpj.ArcomClient) *Buscador {
	return &Buscador{arcom: arcom}
}

// Buscar resolve cada cidade em código de município, busca todas as páginas
// de estabelecimentos para cada CNAE, deduplica por CNPJ (uma empresa pode
// aparecer em mais de um CNAE pesquisado) e calcula os sinais de cada uma.
func (b *Buscador) Buscar(ctx context.Context, cidades []CidadeFiltro, cnaes []string) ([]Prospect, error) {
	cnaesBusca := ExpandirCNAEs(cnaes)

	vistos := map[string]bool{}
	var prospects []Prospect

	for _, cf := range cidades {
		municipio, err := ResolverMunicipio(ctx, b.arcom, cf.Cidade, cf.UF)
		if err != nil {
			return nil, err
		}
		for _, cnae := range cnaesBusca {
			itens, err := b.arcom.EstabelecimentosTodos(ctx, cnpj.EstabelecimentosParams{
				UF:              strings.ToUpper(cf.UF),
				MunicipioCodigo: municipio.Codigo,
				CNAE:            cnae,
			})
			if err != nil {
				return nil, fmt.Errorf("buscar em %s/%s: %w", cf.Cidade, cf.UF, err)
			}
			for _, item := range itens {
				cnpjLimpo := cnpj.Limpar(item.CNPJ)
				if cnpjLimpo == "" || vistos[cnpjLimpo] {
					continue
				}
				vistos[cnpjLimpo] = true
				prospects = append(prospects, mapEstabelecimento(item, cnpjLimpo, municipio.Descricao, strings.ToUpper(cf.UF)))
			}
		}
	}
	return prospects, nil
}

func mapEstabelecimento(e cnpj.EstabelecimentoBruto, cnpjLimpo, cidade, uf string) Prospect {
	logr := juntarNaoVazios(", ", e.Logradouro, e.Numero)
	fim := juntarNaoVazios(", ", e.Bairro, cidade, uf)
	endereco := ""
	if logr != "" || fim != "" {
		endereco = logr
		if fim != "" {
			endereco += " - " + fim
		}
	}
	telefone := ""
	if e.Telefone1 != "" {
		telefone = e.DDD1 + e.Telefone1
	}
	bairro := e.Bairro
	if bairro == "" {
		bairro = "—"
	}

	p := Prospect{
		CNPJ:         cnpjLimpo,
		Razao:        valorOuTraco(e.RazaoSocial),
		NomeFantasia: e.NomeFantasia,
		Atividade:    valorOuTraco(e.CnaeFiscalPrincipal.Descricao),
		CNAECodigo:   e.CnaeFiscalPrincipal.Codigo,
		Bairro:       bairro,
		Endereco:     endereco,
		CEP:          e.CEP,
		Telefone:     telefone,
		Email:        e.CorreioEletronico,
		Porte:        e.PorteEmpresa.Descricao,
		DataInicio:   e.DataInicioAtividade,
		Latitude:     e.Latitude,
		Longitude:    e.Longitude,
		Cidade:       cidade,
		UF:           uf,
	}

	p.Score, p.Temperatura, p.Sinais = scoreLojaFisica(p)
	p.RedeGrande = ehGrandeRede(cnpjLimpo, p.Razao, p.NomeFantasia)
	if meses := mesesDesdeAbertura(p.DataInicio); meses != nil && *meses >= 0 && *meses <= radarMesesNova {
		p.Nova = true
		p.MesesAtivo = meses
	}
	return p
}

func valorOuTraco(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

func juntarNaoVazios(sep string, partes ...string) string {
	var validas []string
	for _, p := range partes {
		if strings.TrimSpace(p) != "" {
			validas = append(validas, p)
		}
	}
	return strings.Join(validas, sep)
}
