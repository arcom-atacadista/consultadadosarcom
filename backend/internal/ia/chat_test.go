package ia

import (
	"context"
	"strings"
	"testing"
)

func TestMontarSystemPrompt(t *testing.T) {
	semContexto := montarSystemPrompt(nil)
	if strings.Contains(semContexto, "já consultou") {
		t.Error("sem empresas de contexto não deveria mencionar 'já consultou'")
	}

	comContexto := montarSystemPrompt([]EmpresaResumo{
		{CNPJ: "11222333000181", Razao: "Empresa Teste", Situacao: "ATIVA", Porte: "DEMAIS", Municipio: "SAO PAULO", UF: "SP", Atividade: "Comércio"},
	})
	if !strings.Contains(comContexto, "Empresa Teste") {
		t.Error("com empresas de contexto deveria citar a razão social")
	}
	if !strings.Contains(comContexto, "buscar_web") {
		t.Error("system prompt deveria mencionar a ferramenta buscar_web")
	}
}

func TestExecutarToolEstatisticasSemPermissao(t *testing.T) {
	chamou := false
	s := &ChatService{
		estatisticas: func(ctx context.Context) (string, error) {
			chamou = true
			return "dados sigilosos", nil
		},
	}
	tc := GroqToolCall{ID: "1"}
	tc.Function.Name = "estatisticas_do_site"

	resultado := s.executarTool(context.Background(), tc, "", false)
	if chamou {
		t.Error("não deveria ter chamado a função de estatísticas pra quem não é admin")
	}
	if strings.Contains(resultado, "dados sigilosos") {
		t.Error("resultado não deveria conter os dados reais pra quem não é admin")
	}
}

func TestExecutarToolEstatisticasComPermissao(t *testing.T) {
	s := &ChatService{
		estatisticas: func(ctx context.Context) (string, error) {
			return "42 contas", nil
		},
	}
	tc := GroqToolCall{ID: "1"}
	tc.Function.Name = "estatisticas_do_site"

	resultado := s.executarTool(context.Background(), tc, "", true)
	if resultado != "42 contas" {
		t.Errorf("resultado = %q, esperado '42 contas'", resultado)
	}
}
