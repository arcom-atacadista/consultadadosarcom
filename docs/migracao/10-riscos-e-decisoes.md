# 10 — Riscos e decisões pendentes

O que o padrão **não** resolve sozinho e depende de gente (infra/gestor). Ficam
isolados aqui para não travar o resto do plano.

## 1. Decisões que precisam de resposta

| # | Decisão | Recomendação | Impacto se adiar |
|---|---------|--------------|------------------|
| 1 | **Host de produção do backend** (servidor, domínio, TLS) — GitHub Pages não roda Go | definir com a infra da plataforma ARCOM antes da Fase 7 | Fases 1–6 sobem só localmente; cutover trava |
| 2 | **Formato de planilha na exportação** (CSV vs XLSX nativo) | CSV no MVP (`encoding/csv`); XLSX só se a infra liberar lib Go no allowlist | export sai como CSV até decidir |
| 3 | **QR code no PDF** é obrigatório? | não no MVP (link de rota basta); se obrigatório, gerar SVG na stdlib | PDF sai com link, sem QR |
| 4 | **Migração de senhas** (Firebase não exporta hash) | usuários fazem reset no 1º acesso, ou admin recria | usuários precisam redefinir senha no cutover |
| 5 | **Migrar histórico de logs?** (`consultas_log`/`atividades_log`) | não migrar (arquivo morto); começar limpo no Postgres | dashboard começa "do zero" pós-cutover |
| 6 | **Presença em Redis vs Postgres** | Redis com TTL (efêmero) | — (recomendação já no plano) |

> As decisões de **conformidade técnica** (Chart.js, export, Firebase, ícones,
> fonte) **já estão resolvidas pelo padrão** em [`08`](08-conformidade-allowlist.md) —
> não são pendências, são fatos do allowlist.

## 2. Riscos

| Risco | Severidade | Mitigação |
|-------|-----------|-----------|
| **Segredos já vazados** no GitHub Pages | Alta | rotacionar as 3 chaves na Fase 0 (não dá pra desvazar) |
| **Perda de paridade** (app tem ~250 funções) | Média | migrar por fases com o app antigo no ar; checklist do guia [`01`](01-diagnostico-atual.md) |
| **Migração de dados** (usuários/pré-cadastros/listas) | Média | export + seed validado, janela de cutover curta |
| **APIs externas mudam de contrato** (Arcom/Trace360) | Média | DTOs isolados por pacote; normalização num ponto só |
| **Limites de terceiros** (429 Groq/Tavily/BrasilAPI) | Baixa | tratamento de 429 no backend + cache |
| **Sem host definido** | Média | Fases 1–6 não dependem de produção; decidir até a 7 |
| **Autorização mal traduzida** (regras Firestore → Go) | Alta | tabela de mapeamento em [`07`](07-seguranca.md); testar cada nível |

## 3. Pré-requisitos antes de codar a Fase 1

1. Decisões #1–#5 desta lista respondidas (ao menos #4 e #5, que afetam a Fase 3).
2. Chaves rotacionadas e guardadas com a infra.
3. Export do Firestore salvo.

## 4. O que este plano NÃO cobre (fora de escopo)

- Provisionamento do ambiente de produção (é da infra da plataforma).
- Custos de terceiros (Groq/Tavily/Trace360) — seguem os mesmos de hoje.
- Novas funcionalidades — o objetivo é **paridade** + conformidade, não features
  novas. Melhorias entram depois do cutover.
