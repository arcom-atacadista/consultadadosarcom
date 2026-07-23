package prospeccao

import (
	"testing"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/cnpj"
)

func TestNomeMunicipioLimpo(t *testing.T) {
	casos := map[string]string{
		"São Simão - GO": "sao simao",
		"São Simão/SP":   "sao simao",
		"Bonito (MS)":    "bonito",
		"Uberlândia":     "uberlandia",
	}
	for entrada, esperado := range casos {
		if got := nomeMunicipioLimpo(entrada); got != esperado {
			t.Errorf("nomeMunicipioLimpo(%q) = %q, esperado %q", entrada, got, esperado)
		}
	}
}

func TestMunicipioBateUF(t *testing.T) {
	casos := []struct {
		nome     string
		m        cnpj.MunicipioBruto
		uf       string
		esperado bool
	}{
		{"uf vazia sempre bate", cnpj.MunicipioBruto{UF: "SP"}, "", true},
		{"uf igual no campo UF", cnpj.MunicipioBruto{UF: "GO"}, "GO", true},
		{"uf diferente no campo UF", cnpj.MunicipioBruto{UF: "SP"}, "GO", false},
		{"uf pela descricao", cnpj.MunicipioBruto{Descricao: "SAO SIMAO - GO"}, "GO", true},
		{"uf pelo prefixo IBGE (GO=52)", cnpj.MunicipioBruto{Codigo: "5219005"}, "GO", true},
		{"uf pelo prefixo IBGE errado", cnpj.MunicipioBruto{Codigo: "3550308"}, "GO", false},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if got := municipioBateUF(c.m, c.uf); got != c.esperado {
				t.Errorf("municipioBateUF() = %v, esperado %v", got, c.esperado)
			}
		})
	}
}
