package cnpj

import "testing"

func TestSimNaoOuTraco(t *testing.T) {
	casos := []struct {
		valor    any
		esperado string
	}{
		{true, "Sim"},
		{false, "Não"},
		{float64(1), "Sim"},
		{float64(0), "Não"},
		{"sim", "Sim"},
		{"true", "Sim"},
		{"1", "Sim"},
		{"s", "Sim"},
		{"nao", "Não"},
		{"", "—"},
		{nil, "—"},
	}
	for _, c := range casos {
		if got := simNaoOuTraco(c.valor); got != c.esperado {
			t.Errorf("simNaoOuTraco(%#v) = %q, esperado %q", c.valor, got, c.esperado)
		}
	}
}

func TestComOuTraco(t *testing.T) {
	if got := comOuTraco(""); got != "—" {
		t.Errorf("comOuTraco(\"\") = %q, esperado —", got)
	}
	if got := comOuTraco("  "); got != "—" {
		t.Errorf("comOuTraco(espaços) = %q, esperado —", got)
	}
	if got := comOuTraco("ativa"); got != "ativa" {
		t.Errorf("comOuTraco(ativa) = %q, esperado ativa", got)
	}
}

func TestJuntarNaoVazios(t *testing.T) {
	got := juntarNaoVazios(", ", "Centro", "", "São Paulo", "SP")
	esperado := "Centro, São Paulo, SP"
	if got != esperado {
		t.Errorf("juntarNaoVazios() = %q, esperado %q", got, esperado)
	}
}

func TestFormatarMilhar(t *testing.T) {
	casos := map[float64]string{
		0:         "0",
		150:       "150",
		1500:      "1.500",
		150000:    "150.000",
		1500000.9: "1.500.000",
	}
	for valor, esperado := range casos {
		if got := formatarMilhar(valor); got != esperado {
			t.Errorf("formatarMilhar(%v) = %q, esperado %q", valor, got, esperado)
		}
	}
}
