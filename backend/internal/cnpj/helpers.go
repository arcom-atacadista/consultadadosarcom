package cnpj

import (
	"strconv"
	"strings"
)

func comOuTraco(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

func strUpper(s string) string {
	return strings.ToUpper(s)
}

// juntarNaoVazios junta as partes não vazias com o separador — equivalente ao
// `.filter(Boolean).join(sep)` do JS usado no app antigo pra montar endereço.
func juntarNaoVazios(sep string, partes ...string) string {
	var validas []string
	for _, p := range partes {
		if strings.TrimSpace(p) != "" {
			validas = append(validas, p)
		}
	}
	return strings.Join(validas, sep)
}

func capitalOuTraco(v *float64) string {
	if v == nil {
		return "—"
	}
	return "R$ " + formatarMilhar(*v)
}

func formatarMilhar(v float64) string {
	inteiro := int64(v)
	s := strconv.FormatInt(inteiro, 10)
	var partes []string
	for len(s) > 3 {
		partes = append([]string{s[len(s)-3:]}, partes...)
		s = s[:len(s)-3]
	}
	partes = append([]string{s}, partes...)
	return strings.Join(partes, ".")
}

// simNaoOuTraco espelha o simNaoOuTraco do app antigo: a API às vezes manda
// bool, às vezes número, às vezes string ("sim"/"true"/"1"/"s").
func simNaoOuTraco(v any) string {
	switch t := v.(type) {
	case bool:
		if t {
			return "Sim"
		}
		return "Não"
	case float64:
		if t != 0 {
			return "Sim"
		}
		return "Não"
	case string:
		s := strings.TrimSpace(strings.ToLower(t))
		if s == "" {
			return "—"
		}
		switch s {
		case "sim", "true", "1", "s":
			return "Sim"
		default:
			return "Não"
		}
	default:
		return "—"
	}
}
