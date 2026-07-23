package prospeccao

import "testing"

func TestScoreLojaFisica(t *testing.T) {
	lat := -23.55
	casos := []struct {
		nome          string
		p             Prospect
		scoreEsperado int
		temp          string
	}{
		{
			nome: "loja de varejo com tudo (quente)",
			p: Prospect{
				CNAECodigo:   "4713",
				NomeFantasia: "Mercadinho do Zé",
				Endereco:     "RUA DAS FLORES, 100 - CENTRO",
				Telefone:     "1133334444", // fixo
				Porte:        "DEMAIS",
				Latitude:     &lat,
				DataInicio:   "2015-01-01", // > 24 meses
			},
			scoreEsperado: 40 + 15 + 15 + 8 + 6 + 8 + 5 + 8, // 105 -> capado em 100
			temp:          "quente",
		},
		{
			nome: "holding/serviço sem nenhum sinal (frio)",
			p: Prospect{
				CNAECodigo: "6821", // imobiliária = serviço/holding
				Endereco:   "S/N",
			},
			scoreEsperado: 0,
			temp:          "frio",
		},
		{
			nome: "endereço sem número não pontua",
			p: Prospect{
				CNAECodigo: "4713",
				Endereco:   "RUA DAS FLORES, S/N - CENTRO",
			},
			scoreEsperado: 40,
			temp:          "morno",
		},
		{
			nome: "telefone celular pontua menos que fixo",
			p: Prospect{
				CNAECodigo: "4713",
				Telefone:   "11987654321", // celular
			},
			scoreEsperado: 40 + 8,
			temp:          "morno",
		},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			score, temp, _ := scoreLojaFisica(c.p)
			if score > 100 {
				score = 100
			}
			if temp != c.temp {
				t.Errorf("temperatura = %q, esperado %q (score=%d)", temp, c.temp, score)
			}
			esperado := c.scoreEsperado
			if esperado > 100 {
				esperado = 100
			}
			if score != esperado {
				t.Errorf("score = %d, esperado %d", score, esperado)
			}
		})
	}
}

func TestScoreLojaFisicaLimites(t *testing.T) {
	// Nunca deve passar de 100 nem ficar negativo.
	p := Prospect{
		CNAECodigo:   "4713",
		NomeFantasia: "Loja",
		Endereco:     "RUA A, 1",
		Telefone:     "1133334444",
		Porte:        "DEMAIS",
		Latitude:     new(float64),
		DataInicio:   "2000-01-01",
	}
	score, _, _ := scoreLojaFisica(p)
	if score < 0 || score > 100 {
		t.Errorf("score fora do intervalo [0,100]: %d", score)
	}
}

func TestMesesDesdeAbertura(t *testing.T) {
	if m := mesesDesdeAbertura(""); m != nil {
		t.Errorf("data vazia deveria dar nil, deu %v", *m)
	}
	if m := mesesDesdeAbertura("data-invalida"); m != nil {
		t.Errorf("data inválida deveria dar nil, deu %v", *m)
	}
	m := mesesDesdeAbertura("2000-01-01")
	if m == nil || *m < 200 {
		t.Errorf("esperado bem mais que 200 meses desde 2000, deu %v", m)
	}
}

func TestTelMovel(t *testing.T) {
	casos := map[string]bool{
		"11987654321": true,  // celular (9 dígitos, começa com 9)
		"1133334444":  false, // fixo
		"119876543":   false, // curto demais
		"":            false,
	}
	for tel, esperado := range casos {
		if got := telMovel(tel); got != esperado {
			t.Errorf("telMovel(%q) = %v, esperado %v", tel, got, esperado)
		}
	}
}
