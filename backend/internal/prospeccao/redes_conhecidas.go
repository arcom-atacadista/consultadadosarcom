package prospeccao

import (
	"regexp"
	"strings"
)

// redesConhecidas é a mesma lista do app antigo: grandes redes/franquias que
// passam pelo corte por nº de filial (cada loja pode ser matriz/filial 1 de
// um CNPJ próprio) — comparação sem acento e por palavra inteira.
var redesConhecidas = []string{
	// farmácias / drogarias
	"pague menos", "paguemenos", "pague-menos", "ultrapopular", "ultrafarma", "drogasil", "droga raia",
	"drogaraia", "raia", "drogaria sao paulo", "drogarias pacheco", "pacheco", "venancio", "panvel",
	"extrafarma", "nissei", "drogaria araujo", "araujo", "bifarma", "onofre", "drogal", "drogao",
	"indiana", "sao joao", "drogaria sao joao", "pague", "drogarias globo", "globo",
	// perfumaria / cosméticos
	"o boticario", "boticario", "natura", "jequiti", "eudora", "quem disse berenice", "mahogany",
	"avon", "the beauty box",
	// supermercados / atacado
	"carrefour", "atacadao", "assai", "pao de acucar", "bompreco", "sams club", "makro",
	"tenda atacado", "grupo mateus", "gbarbosa", "angeloni", "condor", "muffato", "zaffari",
	"savegnago", "sonda", "st marche", "supermercados bh", "super nosso", "verdemar", "epa", "dma",
	"tonin", "cencosud", "bh supermercados", "apoio mineiro", "villefort", "coop",
	// varejo geral / franquias
	"magazine luiza", "magalu", "casas bahia", "ponto frio", "americanas", "lojas americanas",
	"riachuelo", "lojas renner", "renner", "marisa", "pernambucanas", "havan", "polishop",
	"kalunga", "leroy merlin", "telhanorte", "cobasi", "petz", "cacau show", "kopenhagen",
	"ri happy", "centauro",
	// fast food / franquias de alimentação
	"mcdonalds", "burger king", "subway", "outback", "habibs", "giraffas", "pizza hut", "spoleto", "bobs",
}

var acentoReplacer = strings.NewReplacer(
	"á", "a", "à", "a", "ã", "a", "â", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "õ", "o", "ô", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c",
)

// semAcento equivale ao semAcento(s) do app antigo: minúsculo e sem acento.
func semAcento(s string) string {
	return acentoReplacer.Replace(strings.ToLower(s))
}

var redesRegex = compilarRedesRegex()

func compilarRedesRegex() *regexp.Regexp {
	partes := make([]string, len(redesConhecidas))
	for i, r := range redesConhecidas {
		partes[i] = regexp.QuoteMeta(semAcento(r))
	}
	return regexp.MustCompile(`\b(` + strings.Join(partes, "|") + `)\b`)
}

func eRedeConhecida(razao, nomeFantasia string) bool {
	texto := semAcento(razao + " " + nomeFantasia)
	return redesRegex.MatchString(texto)
}

// ordemFilial extrai os 4 dígitos do meio do CNPJ (número da filial — 0001 é
// matriz; um número alto indica rede grande, ex.: Americanas filial 0842).
func ordemFilial(cnpjLimpo string) int {
	if len(cnpjLimpo) != 14 {
		return 1
	}
	n := 0
	for _, r := range cnpjLimpo[8:12] {
		n = n*10 + int(r-'0')
	}
	if n == 0 {
		return 1
	}
	return n
}

const limiteFiliaisPadrao = 20

// ehGrandeRede junta os dois sinais do app antigo: filial alta (>= limite)
// OU nome bate numa rede/franquia conhecida.
func ehGrandeRede(cnpjLimpo, razao, nomeFantasia string) bool {
	return ordemFilial(cnpjLimpo) >= limiteFiliaisPadrao || eRedeConhecida(razao, nomeFantasia)
}
