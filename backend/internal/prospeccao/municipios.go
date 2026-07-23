package prospeccao

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/cnpj"
)

// ibgeUF mapeia sigla de UF -> prefixo do código IBGE do município (2
// primeiros dígitos), usado como último critério de desempate.
var ibgeUF = map[string]string{
	"RO": "11", "AC": "12", "AM": "13", "RR": "14", "PA": "15", "AP": "16", "TO": "17",
	"MA": "21", "PI": "22", "CE": "23", "RN": "24", "PB": "25", "PE": "26", "AL": "27",
	"SE": "28", "BA": "29", "MG": "31", "ES": "32", "RJ": "33", "SP": "35", "PR": "41",
	"SC": "42", "RS": "43", "MS": "50", "MT": "51", "GO": "52", "DF": "53",
}

var reSufixoUF = regexp.MustCompile(`[\s/(\-]+[A-Za-z]{2}\)?\s*$`)

// nomeMunicipioLimpo tira acento e o sufixo de UF (" - GO", "/GO", " (GO)").
func nomeMunicipioLimpo(desc string) string {
	s := reSufixoUF.ReplaceAllString(desc, "")
	return semAcento(strings.TrimSpace(s))
}

// municipioBateUF confere se o candidato de município é da UF pedida —
// existem cidades homônimas em estados diferentes (ex.: São Simão em GO e SP).
func municipioBateUF(m cnpj.MunicipioBruto, uf string) bool {
	if uf == "" {
		return true
	}
	alvo := strings.ToUpper(uf)
	if strings.ToUpper(m.UF) == alvo {
		return true
	}
	desc := strings.ToUpper(m.Descricao)
	if matched, _ := regexp.MatchString(`[\s/\-(]`+alvo+`\b`, desc); matched {
		return true
	}
	if prefixo, ok := ibgeUF[alvo]; ok && strings.HasPrefix(m.Codigo, prefixo) {
		return true
	}
	return false
}

// ResolverMunicipio resolve o nome da cidade digitado num código de
// município da Receita Federal, respeitando a UF (docs/migracao/01 §3.1).
func ResolverMunicipio(ctx context.Context, arcom *cnpj.ArcomClient, nomeCidade, uf string) (*cnpj.MunicipioBruto, error) {
	lista, err := arcom.Municipios(ctx, nomeCidade)
	if err != nil {
		return nil, err
	}
	if len(lista) == 0 {
		return nil, fmt.Errorf("município %q não encontrado na Receita Federal", nomeCidade)
	}

	naUF := lista
	if uf != "" {
		naUF = nil
		for _, m := range lista {
			if municipioBateUF(m, uf) {
				naUF = append(naUF, m)
			}
		}
		if len(naUF) == 0 {
			naUF = lista
		}
	}

	alvo := semAcento(nomeCidade)
	for i := range naUF {
		if nomeMunicipioLimpo(naUF[i].Descricao) == alvo {
			return &naUF[i], nil
		}
	}
	return &naUF[0], nil
}
