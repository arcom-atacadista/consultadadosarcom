package ia

import "testing"

func TestSemTraco(t *testing.T) {
	casos := map[string]string{
		"":            "",
		"—":           "",
		"11912345678": "11912345678",
	}
	for in, esperado := range casos {
		if got := semTraco(in); got != esperado {
			t.Errorf("semTraco(%q) = %q, esperado %q", in, got, esperado)
		}
	}
}

func TestValorOuPadrao(t *testing.T) {
	if got := valorOuPadrao("", "padrão"); got != "padrão" {
		t.Errorf("valorOuPadrao vazio = %q, esperado padrão", got)
	}
	if got := valorOuPadrao("  ", "padrão"); got != "padrão" {
		t.Errorf("valorOuPadrao espaços = %q, esperado padrão", got)
	}
	if got := valorOuPadrao("texto", "padrão"); got != "texto" {
		t.Errorf("valorOuPadrao com texto = %q, esperado texto", got)
	}
}

func TestStrFromPtr(t *testing.T) {
	if got := strFromPtr(nil); got != "" {
		t.Errorf("strFromPtr(nil) = %q, esperado vazio", got)
	}
	s := "valor"
	if got := strFromPtr(&s); got != "valor" {
		t.Errorf("strFromPtr(&s) = %q, esperado valor", got)
	}
}
