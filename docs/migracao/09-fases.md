# 09 — Fases da migração

Migração incremental (sem "big bang"). Cada fase é um conjunto pequeno de commits
verificáveis na branch `claude/arcom-ai-first-migration-qvpj17`, com **critério
de aceite** claro. O `index.html` atual segue no ar (GitHub Pages) até o cutover
da Fase 7.

---

## Fase 0 — Preparação (sem código de app)
**Objetivo:** destravar decisões e segurança.
- Fechar as decisões de [`10`](10-riscos-e-decisoes.md) com infra/gestor
  (host de produção, planilha CSV vs XLSX, QR, migração de senhas).
- **Rotacionar** `ARCOM_API_KEY`, `TRACE360_API_KEY`, `GEOAPIFY_API_KEY`.
- Exportar snapshot do Firestore (backup + base do seed).
**Aceite:** chaves novas guardadas com a infra; export do Firestore salvo.

## Fase 1 — Fundação (esqueleto que sobe)
**Objetivo:** `docker compose up` respondendo.
- Trazer a base da spec: `docker-compose.yml`, `.env.example`, `.gitignore`,
  `frontend/` (Vite + Tailwind + tokens ARCOM + `.npmrc` + proxy) e `backend/`
  (Go + chi + `GET /api/health`).
- Adicionar as variáveis de ambiente das chaves no serviço `backend`.
**Aceite:** `docker compose up --build` abre em `http://localhost:8080`;
`GET /api/health` → 200; healthcheck do backend verde. Sem função de negócio.

## Fase 2 — Design System + shell da UI
**Objetivo:** identidade ARCOM e navegação.
- Red Hat Display no `index.css`; componentes shadcn com tokens (Button, Card,
  Badge, Input, Select, Alert, Dialog, Table, Tabs).
- Shell: Sidebar (`verde-escuro`) + Topbar + layout `bg-surface`; ícones
  `lucide-react`; rotas das 5 abas + Login/Cover/Pendente (telas vazias).
**Aceite:** navegação entre telas vazias com a cara da ARCOM; sem emoji; nenhuma
cor/fonte fora dos tokens.

## Fase 3 — Auth (remove o Firebase do caminho)
**Objetivo:** login próprio, sem Firebase.
- Backend `auth`/`usuarios`: register (pendente), login (JWT/bcrypt), `me`,
  troca de senha, aprovação/promoção; tabela `usuarios`; super-admin por e-mail.
- Frontend: telas de login/cadastro/pendente, guarda de rota, token no axios.
- Seed dos usuários **aprovados** do Firestore (senhas via reset — ver [`10`]).
**Aceite:** login funciona; conta nova entra pendente; admin aprova; rotas
protegidas por status/admin.

## Fase 4 — Consulta de CNPJ (primeiro módulo de valor)
**Objetivo:** paridade da aba Consulta.
- Backend `cnpj`: token Arcom server-side, `POST /v1/cnpj/batch` (chunk 1000),
  fallback BrasilAPI, cache 24h (Redis), `consultas_log`, validação de dígito.
- Frontend: consulta individual/lote, seletor de fonte, "Cliente Arcom",
  favoritos/histórico (localStorage), atalhos de teclado.
**Aceite:** consulta individual e lote retornam o mesmo conteúdo do app antigo;
reconsulta em 24h não gasta chamada; chave Arcom **não** aparece no navegador.

## Fase 5 — Prospecção + Loja física + Pré-cadastro
**Objetivo:** paridade da aba Prospecção.
- Backend `prospeccao`+`geo`: `municipios`, `estabelecimentos` (paginado, dedup),
  corte de grandes redes, **score de loja física**, `listas_prospeccao`,
  `pre_cadastros`, geocode Geoapify.
- Frontend: filtros multi-cidade/ramo, tabela ordenada por score, semáforo por
  bolinhas (tokens), modal Fachada/Mapa (Street View embed), contatos clicáveis,
  pré-cadastro, ranking por bairro.
- **PDF de indicação** via HTML de impressão + print-to-PDF ([`08`](08-conformidade-allowlist.md) §3).
**Aceite:** busca multi-cidade/ramo com dedup; score idêntico à regra do app
antigo; PDF sai agrupado por ramo com telefone e link de rota.

## Fase 6 — IA + Enriquecimento Trace360
**Objetivo:** paridade dos recursos de IA e do dossiê.
- Backend `ia`: Groq (insight/ranking/chat com tool-calling) + Tavily
  server-side (substitui o Worker); ferramentas `buscar_web` e
  `estatisticas_do_site` (com permissão).
- Backend `enriquecimento`: `POST /clientes` na Trace360 (x-api-key server-side),
  posse por usuário em tabela, status assíncrono via `asynq`/`cron`.
- Frontend: insight por empresa e em lote, ranking por IA, chat com histórico,
  aba Enriquecimento com progresso por CNPJ.
**Aceite:** insight/chat funcionam sem o Worker; nenhuma chave de IA no
navegador; enriquecimento em lote (até 500) com status atualizando.

## Fase 7 — Conversão + Admin + acabamento + cutover
**Objetivo:** fechar paridade e virar o tráfego.
- Backend `conversao` (planilha de vendas + retorno) e `admin` (dashboard,
  atividades, presença via Redis).
- Frontend: aba Conversão; dashboard admin com **gráficos SVG/Tailwind**
  ([`08`](08-conformidade-allowlist.md) §2); atividade por usuário; exportação
  (CSV/JSON/PDF); tema claro/escuro + cor de destaque + partículas; PWA;
  acessibilidade (`prefers-reduced-motion`).
- **Cutover:** migrar dados vivos ([`04`](04-modelo-dados.md) §4), apontar
  produção para o Docker novo, aposentar GitHub Pages e o projeto Firebase,
  remover `index.html`/`firestore.rules` da raiz.
**Aceite:** todas as 5 abas + IA com paridade; dashboard sem Chart.js; produção
no novo deploy; Firebase desligado; segredos só no backend.

---

## Ordem e dependências

```
0 ─▶ 1 ─▶ 2 ─▶ 3 ─▶ 4 ─▶ 5 ─▶ 6 ─▶ 7
             (3 destrava tudo que exige login)
```

Fase 3 é o marco de segurança (mata o Firebase e a maior superfície de risco).
Fases 4–7 entregam módulo a módulo, sempre com o app antigo como rede de
segurança até o cutover.
