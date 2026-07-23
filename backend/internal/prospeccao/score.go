package prospeccao

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var reTemDigito = regexp.MustCompile(`\d`)
var reSemNumero = regexp.MustCompile(`(?i)\bS/?N\b`)

var cnaesIndustria = map[string]bool{
	"10": true, "11": true, "13": true, "14": true, "15": true, "16": true, "17": true,
	"18": true, "20": true, "22": true, "23": true, "25": true, "31": true, "32": true, "33": true,
}
var cnaesServicoHolding = map[string]bool{
	"64": true, "65": true, "66": true, "68": true, "69": true, "70": true, "82": true,
}

// telMovel confere se o telefone (com DDD) parece celular: 9 dígitos após o
// DDD, começando com 9. Só nesse caso faz sentido oferecer WhatsApp.
func telMovel(tel string) bool {
	d := apenasDigitos(tel)
	if len(d) < 11 {
		return false
	}
	num := d[2:]
	return len(num) == 9 && num[0] == '9'
}

func apenasDigitos(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// mesesDesdeAbertura conta os meses entre a data ISO de abertura e hoje.
func mesesDesdeAbertura(dataISO string) *int {
	if len(dataISO) < 10 {
		return nil
	}
	abertura, err := time.Parse("2006-01-02", dataISO[:10])
	if err != nil {
		return nil
	}
	hoje := time.Now()
	meses := (hoje.Year()-abertura.Year())*12 + int(hoje.Month()) - int(abertura.Month())
	return &meses
}

// scoreLojaFisica reproduz fielmente docs/migracao/01-diagnostico-atual.md §6
// (mesma regra do app antigo, agora calculada no backend).
func scoreLojaFisica(p Prospect) (score int, temperatura string, sinais []string) {
	p2 := p.CNAECodigo
	if len(p2) > 2 {
		p2 = p2[:2]
	}

	var cnaeScore int
	var cnaeLabel string
	switch {
	case p2 == "47":
		cnaeScore, cnaeLabel = 40, "Varejo"
	case p2 == "56":
		cnaeScore, cnaeLabel = 26, "Alimentação"
	case p2 == "45":
		cnaeScore, cnaeLabel = 24, "Veículos"
	case p2 == "86":
		cnaeScore, cnaeLabel = 20, "Saúde"
	case cnaesIndustria[p2]:
		cnaeScore, cnaeLabel = 14, "Indústria"
	case cnaesServicoHolding[p2]:
		cnaeScore, cnaeLabel = 0, "Serviço/holding"
	default:
		cnaeScore, cnaeLabel = 12, "Outro"
	}
	score += cnaeScore
	if cnaeScore >= 24 {
		sinais = append(sinais, cnaeLabel)
	}

	if strings.TrimSpace(p.NomeFantasia) != "" {
		score += 15
		sinais = append(sinais, "Nome fantasia")
	}

	logradouro := p.Endereco
	if i := strings.Index(logradouro, " - "); i >= 0 {
		logradouro = logradouro[:i]
	}
	temNumero := reTemDigito.MatchString(logradouro) && !reSemNumero.MatchString(p.Endereco)
	if temNumero {
		score += 15
		sinais = append(sinais, "Endereço c/ número")
	}

	tel := apenasDigitos(p.Telefone)
	if tel != "" {
		score += 8
		if !telMovel(tel) {
			score += 6
			sinais = append(sinais, "Tel fixo")
		} else {
			sinais = append(sinais, "Telefone")
		}
	}

	if p.Porte != "" && p.Porte != "MICRO EMPRESA" && p.Porte != "—" {
		score += 8
	}
	if p.Latitude != nil {
		score += 5
	}

	meses := mesesDesdeAbertura(p.DataInicio)
	if meses != nil && *meses >= 24 {
		score += 8
		sinais = append(sinais, fmt.Sprintf("%d anos", *meses/12))
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	switch {
	case score >= 65:
		temperatura = "quente"
	case score >= 40:
		temperatura = "morno"
	default:
		temperatura = "frio"
	}
	return score, temperatura, sinais
}
