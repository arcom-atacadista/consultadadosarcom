package cnpj

import "testing"

func TestValido(t *testing.T) {
	casos := []struct {
		nome  string
		cnpj  string
		valid bool
	}{
		{"formatado com máscara", "11.222.333/0001-81", true},
		{"só dígitos", "11222333000181", true},
		{"petrobras (real)", "33.000.167/0001-01", true},
		{"dígito verificador errado", "11222333000199", false},
		{"todos os dígitos iguais", "00000000000000", false},
		{"muito curto", "1122233300018", false},
		{"muito longo", "112223330001811", false},
		{"vazio", "", false},
		{"letras", "abcdefghijklmn", false},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if got := Valido(c.cnpj); got != c.valid {
				t.Errorf("Valido(%q) = %v, esperado %v", c.cnpj, got, c.valid)
			}
		})
	}
}

func TestLimpar(t *testing.T) {
	if got := Limpar("11.222.333/0001-81"); got != "11222333000181" {
		t.Errorf("Limpar() = %q, esperado 11222333000181", got)
	}
}
