package cnpj

import "testing"

func TestDividirEmChunks(t *testing.T) {
	casos := []struct {
		nome     string
		total    int
		tamanho  int
		esperado []int // tamanho de cada chunk esperado
	}{
		{"vazio", 0, 1000, nil},
		{"menor que o tamanho do chunk", 5, 1000, []int{5}},
		{"exatamente 1 chunk", 1000, 1000, []int{1000}},
		{"passa de 1 chunk por 1", 1001, 1000, []int{1000, 1}},
		{"1500 em chunks de 1000", 1500, 1000, []int{1000, 500}},
		{"2500 em chunks de 1000", 2500, 1000, []int{1000, 1000, 500}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			cnpjs := make([]string, c.total)
			for i := range cnpjs {
				cnpjs[i] = "cnpj"
			}
			chunks := dividirEmChunks(cnpjs, c.tamanho)
			if len(chunks) != len(c.esperado) {
				t.Fatalf("got %d chunks, esperado %d", len(chunks), len(c.esperado))
			}
			total := 0
			for i, chunk := range chunks {
				if len(chunk) != c.esperado[i] {
					t.Errorf("chunk[%d] tem %d itens, esperado %d", i, len(chunk), c.esperado[i])
				}
				total += len(chunk)
			}
			if total != c.total {
				t.Errorf("total de itens nos chunks = %d, esperado %d", total, c.total)
			}
		})
	}
}
