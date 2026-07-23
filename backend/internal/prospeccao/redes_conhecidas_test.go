package prospeccao

import "testing"

func TestEhGrandeRede(t *testing.T) {
	casos := []struct {
		nome         string
		cnpj         string
		razao        string
		nomeFantasia string
		esperado     bool
	}{
		{"matriz de empresa pequena", "11222333000181", "Mercadinho do Bairro Ltda", "", false},
		{"filial alta (>=20)", "11222333002599", "Empresa Qualquer", "", true},
		{"nome bate rede conhecida (razão)", "11222333000181", "Carrefour Comercio Ltda", "", true},
		{"nome bate rede conhecida (fantasia)", "11222333000181", "Comercial XYZ", "Droga Raia", true},
		{"nome parecido mas não é a rede (palavra inteira)", "11222333000181", "Raiaria Materiais", "", false},
		{"filial 0001 (matriz) não é rede grande", "11222333000181", "Empresa Comum", "", false},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if got := ehGrandeRede(c.cnpj, c.razao, c.nomeFantasia); got != c.esperado {
				t.Errorf("ehGrandeRede() = %v, esperado %v", got, c.esperado)
			}
		})
	}
}

func TestOrdemFilial(t *testing.T) {
	casos := map[string]int{
		"11222333000181": 1,
		"11222333084200": 842,
		"curto":          1,
	}
	for cnpj, esperado := range casos {
		if got := ordemFilial(cnpj); got != esperado {
			t.Errorf("ordemFilial(%q) = %d, esperado %d", cnpj, got, esperado)
		}
	}
}

func TestSemAcento(t *testing.T) {
	if got := semAcento("São João"); got != "sao joao" {
		t.Errorf("semAcento() = %q, esperado 'sao joao'", got)
	}
}

func TestExpandirCNAEs(t *testing.T) {
	out := ExpandirCNAEs([]string{"MISTO", "4713", "4711302"}) // 4711302 já está no MISTO
	if len(out) != len(cnaesMisto)+1 {
		t.Errorf("esperado %d cnaes (misto + 4713 sem duplicar 4711302), deu %d: %v", len(cnaesMisto)+1, len(out), out)
	}
}
