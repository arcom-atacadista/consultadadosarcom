package cnpj

import "regexp"

var apenasDigitos = regexp.MustCompile(`\D`)

// Limpar remove tudo que não é dígito (pontos, barra, hífen).
func Limpar(cnpj string) string {
	return apenasDigitos.ReplaceAllString(cnpj, "")
}

// Valido confere o dígito verificador do CNPJ — nunca gastamos uma chamada
// de API externa com um CNPJ que já sabemos que é inválido (padroes/03).
func Valido(cnpj string) bool {
	c := Limpar(cnpj)
	if len(c) != 14 {
		return false
	}
	if todosIguais(c) {
		return false
	}
	digitos := make([]int, 14)
	for i, r := range c {
		digitos[i] = int(r - '0')
	}
	d1 := digitoVerificador(digitos[:12], []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	if d1 != digitos[12] {
		return false
	}
	d2 := digitoVerificador(digitos[:13], []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	return d2 == digitos[13]
}

func digitoVerificador(nums []int, pesos []int) int {
	soma := 0
	for i, n := range nums {
		soma += n * pesos[i]
	}
	resto := soma % 11
	if resto < 2 {
		return 0
	}
	return 11 - resto
}

func todosIguais(s string) bool {
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}
