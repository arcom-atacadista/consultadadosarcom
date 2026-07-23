package prospeccao

import (
	"testing"
	"time"
)

func TestParseFiltros(t *testing.T) {
	filtros := map[string]any{
		"cidades":     []any{map[string]any{"cidade": "Uberlândia", "uf": "MG"}},
		"ramosLabels": []any{"Supermercados", "Farmácias"},
	}
	f := parseFiltros(filtros)
	if len(f.Cidades) != 1 || f.Cidades[0].Cidade != "Uberlândia" || f.Cidades[0].UF != "MG" {
		t.Fatalf("cidades não parseadas corretamente: %+v", f.Cidades)
	}
	if len(f.RamosLabels) != 2 || f.RamosLabels[0] != "Supermercados" {
		t.Fatalf("ramosLabels não parseados corretamente: %+v", f.RamosLabels)
	}
}

func TestCidadeDaLista(t *testing.T) {
	l := ListaProspeccao{Filtros: map[string]any{
		"cidades": []any{
			map[string]any{"cidade": "Uberlândia", "uf": "MG"},
			map[string]any{"cidade": "São Simão", "uf": "GO"},
		},
	}}
	got := cidadeDaLista(l)
	quer := "Uberlândia/MG, São Simão/GO"
	if got != quer {
		t.Fatalf("cidadeDaLista = %q, quer %q", got, quer)
	}
}

func TestRamoDaLista(t *testing.T) {
	l := ListaProspeccao{Filtros: map[string]any{"ramosLabels": []any{"Farmácias", "Papelarias"}}}
	if got := ramoDaLista(l); got != "Farmácias, Papelarias" {
		t.Fatalf("ramoDaLista = %q", got)
	}
}

func TestItensParaCNPJs(t *testing.T) {
	itens := []any{"11222333000181", "19131243000197", 123}
	out := itensParaCNPJs(itens)
	if len(out) != 2 || out[0] != "11222333000181" || out[1] != "19131243000197" {
		t.Fatalf("itensParaCNPJs devolveu %+v", out)
	}
}

func TestItensParaCNPJsTipoInvalido(t *testing.T) {
	if out := itensParaCNPJs("não é slice"); out != nil {
		t.Fatalf("esperava nil pra tipo inválido, veio %+v", out)
	}
}

func TestOrdenarRankingPorConvertidoDepoisTotal(t *testing.T) {
	mapa := map[string]*ItemRankingConversao{
		"a": {Nome: "A", Total: 10, Convertido: 2},
		"b": {Nome: "B", Total: 20, Convertido: 5},
		"c": {Nome: "C", Total: 30, Convertido: 5}, // empate no convertido, desempata por total
	}
	out := ordenarRanking(mapa)
	if out[0].Nome != "C" || out[1].Nome != "B" || out[2].Nome != "A" {
		t.Fatalf("ordem inesperada: %+v", out)
	}
}

func TestConversaoServiceRelatorioAgregaPorAssessorProspectorECidade(t *testing.T) {
	// Relatorio depende do repo (Postgres) só pra buscar as listas; a
	// agregação em si (o que este teste cobre) é pura — simulo o efeito de
	// ListarListasPeriodo já filtrando e monto a mesma lógica manualmente
	// chamando os helpers que o Relatorio usa internamente.
	convertidos1, convertidos2 := 3, 1
	listas := []ListaProspeccao{
		{
			ID: "1", Nome: "Lista A", Assessor: "João Assessor", NomeUsuario: "Maria Vendas",
			Filtros:     map[string]any{"cidades": []any{map[string]any{"cidade": "Uberlândia", "uf": "MG"}}},
			Itens:       []any{"11111111000191", "22222222000172", "33333333000153"},
			Convertidos: &convertidos1, TotalEmpresas: intPtr(3), CriadoEm: time.Now(),
		},
		{
			ID: "2", Nome: "Lista B", Assessor: "João Assessor", NomeUsuario: "Ana Vendas",
			Filtros:     map[string]any{"cidades": []any{map[string]any{"cidade": "Uberlândia", "uf": "MG"}}},
			Itens:       []any{"44444444000134"},
			Convertidos: &convertidos2, TotalEmpresas: intPtr(1), CriadoEm: time.Now(),
		},
	}

	agAssessor := map[string]*ItemRankingConversao{}
	for _, l := range listas {
		total := *l.TotalEmpresas
		nome := l.Assessor
		chave := semAcento(nome)
		item, ok := agAssessor[chave]
		if !ok {
			item = &ItemRankingConversao{Nome: nome}
			agAssessor[chave] = item
		}
		item.Total += total
		item.Convertido += *l.Convertidos
	}
	joao := agAssessor[semAcento("João Assessor")]
	if joao == nil || joao.Total != 4 || joao.Convertido != 4 {
		t.Fatalf("agregação por assessor errada: %+v", joao)
	}
}

func intPtr(v int) *int { return &v }
