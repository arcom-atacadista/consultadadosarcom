package admin

import "testing"

func TestCnpjsDoPayload(t *testing.T) {
	payload := map[string]any{"cnpjs": []any{"11111111000191", "22222222000172", 42}}
	out := cnpjsDoPayload(payload)
	if len(out) != 2 || out[0] != "11111111000191" || out[1] != "22222222000172" {
		t.Fatalf("cnpjsDoPayload devolveu %+v", out)
	}
}

func TestCnpjsDoPayloadTipoInvalido(t *testing.T) {
	if out := cnpjsDoPayload("não é map"); out != nil {
		t.Fatalf("esperava nil, veio %+v", out)
	}
	if out := cnpjsDoPayload(map[string]any{"outraChave": 1}); out != nil {
		t.Fatalf("esperava nil sem a chave cnpjs, veio %+v", out)
	}
}
