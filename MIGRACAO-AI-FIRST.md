# Plano de Migração — CDA para o Padrão ARCOM AI-First (v2.1)

> **Documento de planejamento.** Descreve como migrar o app atual
> (`index.html` monolítico no GitHub Pages) para a arquitetura AI-First da
> ARCOM: monorepo `frontend/` (React + Vite + Tailwind + shadcn) + `backend/`
> (Go + chi) subindo com `docker compose up --build`.
>
> Fonte da verdade do padrão: pasta `padroes/` da spec (v2.1). Este plano
> **não altera** o app atual — organiza a reconstrução.

---

## 1. Resumo executivo

O **CDA (Consulta Dados Arcom)** hoje é um único `index.html` de ~6.900 linhas
(366 KB), HTML/CSS/JS puro, hospedado no GitHub Pages, que fala **direto do
navegador** com Firebase (Auth + Firestore) e várias APIs — algumas com
**chave secreta embutida no código do frontend**.

O padrão ARCOM AI-First exige o oposto: frontend React só desenha a tela e chama
`/api/...`; toda lógica, segredo e acesso a dados fica num **backend Go**; sobe
tudo com um comando via Docker. Portanto a migração é uma **reescrita
estruturada**, não um ajuste. Ela resolve de quebra o problema mais grave de
hoje: **segredos expostos no bundle**.

**Esforço estimado:** grande (o app tem 5 módulos ricos e ~250 funções). O plano
divide em **7 fases** que entregam valor incrementalmente, começando pela
fundação e pela remoção dos segredos.

---

## 2. Diagnóstico do app atual

### 2.1 O que o app faz (5 abas)

| Aba | Função | Depende de |
|-----|--------|-----------|
| **Consulta** | Consulta CNPJ (individual/lote até 1000), cache 24h, valida dígito, favoritos, histórico, já-é-cliente-Arcom | API Consulta CNPJ Arcom, BrasilAPI |
| **Prospecção** | Busca empresas ativas não-clientes por cidade/UF/CNAE, corte de grandes redes, score "provável loja física" (0–100) + Street View, ranking por bairro, pré-cadastro, PDF de indicação com rota | API Arcom, Geoapify (geocode), Google Maps/Street View |
| **Enriquecimento** | Dossiê Trace360 por CNPJ (fluxo assíncrono com progresso) | API Trace360 (`trace360ai.arcom.com.br`) |
| **Conversão** | Controle de conversão da prospecção (retorno + assessor), processa planilha de vendas | Firestore |
| **Admin** | Dashboard multiusuário: contas, logins, consultas, prospecções, PDFs, histórico de atividades, presença online | Firebase Auth + Firestore, Chart.js |

Recursos transversais: **Insight Comercial com IA** (Groq Llama 3.3 70B + Tavily
busca web) e **chat** (Groq 8B com tool-calling); exportação Excel/CSV/JSON/PDF;
tema claro/escuro, cor de destaque, partículas; PWA-like.

### 2.2 Stack atual vs. allowlist ARCOM

| Item hoje | Situação no padrão | Ação |
|-----------|--------------------|------|
| HTML/CSS/JS vanilla, sem build | ❌ | → React + Vite + Tailwind + shadcn |
| Firebase Auth | ❌ (fora do allowlist) | → Go + `golang-jwt` + `bcrypt` + Postgres |
| Firestore (7 coleções) | ❌ (fora do allowlist) | → Postgres (`gorm`/`pgx` + `goose`) |
| Font Awesome (CDN) | ❌ | → `lucide-react` |
| Chart.js (CDN) | ⚠️ **não está no allowlist** | ver §6 (decisão) |
| SheetJS / html2pdf / qrcodejs (CDN) | ⚠️ export não coberto pelo allowlist | ver §6 (decisão) |
| Google Fonts CDN | ⚠️ | mantido via `@import` no CSS (o design-system permite) |
| Chamadas diretas a API externa do navegador | ❌ proibido | → tudo via backend Go |
| **Segredos no frontend** | ❌ proibido | → variáveis de ambiente no backend |
| Cloudflare Worker proxy (Groq/Tavily) | substituível | → endpoints no backend Go |
| GitHub Pages | ❌ (não roda backend) | → Docker (infra da plataforma) |

### 2.3 Segredos expostos hoje (corrigir na Fase 1)

Estão **em texto plano no `index.html`** e vão para qualquer navegador:

- `TRACE360_API_KEY` (linha ~2447) — API Trace360
- `ARCOM_API_KEY` (linha ~3842) — API Consulta CNPJ Arcom
- `GEOAPIFY_API_KEY` (linha ~5855) — geocodificação
- `firebaseConfig.apiKey` (linha ~3003) — pública por design do Firebase, mas
  sai de cena com a migração
- Groq/Tavily: já ficam atrás do Worker `consultarcom.jlrodrigues6900.workers.dev`
  (as chaves não estão no bundle) — migram para o backend Go.

> **Ação imediata recomendada** (independente da migração): rotacionar
> `TRACE360_API_KEY`, `ARCOM_API_KEY` e `GEOAPIFY_API_KEY`, pois já estão
> publicadas no histórico do GitHub Pages.

---

## 3. Arquitetura alvo

```
consultadadosarcom/
├── docker-compose.yml          # front + back + postgres + redis (base da spec)
├── .env.example
├── README.md
│
├── frontend/                   # React + Vite (só desenha e chama /api)
│   ├── .npmrc  Dockerfile  vite.config.ts  tailwind.config.js  postcss.config.js
│   ├── package.json  pnpm-lock.yaml  index.html
│   └── src/
│       ├── main.tsx  App.tsx
│       ├── lib/{api.ts,cn.ts}
│       ├── components/ui/       # shadcn com tokens ARCOM (Button, Card, ...)
│       ├── components/          # componentes do domínio
│       ├── pages/               # 1 por aba: Consulta, Prospeccao, Enriquecimento, Conversao, Admin, Login
│       └── hooks/
│
└── backend/                    # Go (toda lógica, segredos e dados)
    ├── Dockerfile  go.mod  go.sum  .env.example
    ├── cmd/server/main.go       # router chi + GET /api/health
    └── internal/
        ├── config/  db/  redis/  http/
        ├── auth/                # login/JWT/bcrypt, aprovação de contas
        ├── cnpj/                # proxy Consulta CNPJ Arcom + BrasilAPI + cache
        ├── prospeccao/          # busca, score loja física, listas salvas
        ├── enriquecimento/      # proxy Trace360 (assíncrono)
        ├── geo/                 # proxy Geoapify
        ├── ia/                  # proxy Groq + Tavily (insight, ranking, chat)
        ├── conversao/           # controle de conversão + planilha de vendas
        └── admin/               # dashboard, atividades, presença
```

**Regra de ouro:** navegador → `frontend:8080`; `/api/*` → proxy Vite →
`backend:3000`; backend → Postgres/Redis + APIs externas. Uma origem só, sem
CORS, segredo nenhum no navegador.

---

## 4. Mapa de migração: browser → backend

Cada chamada externa que hoje sai do navegador vira um endpoint `/api`:

| Hoje (no navegador) | Vira endpoint backend | Pacote Go |
|---------------------|----------------------|-----------|
| Firebase Auth (login/senha) | `POST /api/auth/login`, `/register`, `/me`, `/senha` | `auth` |
| Firestore `usuarios` | `GET/PATCH /api/usuarios` (aprovar/promover) | `auth`/`admin` |
| API Consulta CNPJ Arcom (JWT + lote) | `POST /api/cnpj/consultar` (lote), cache no backend | `cnpj` |
| BrasilAPI | fallback interno de `/api/cnpj/...` | `cnpj` |
| Busca de prospecção (API Arcom) | `POST /api/prospeccao/buscar` | `prospeccao` |
| Geoapify geocode | `GET /api/geo/geocode?q=` | `geo` |
| Groq (insight/ranking/chat) via Worker | `POST /api/ia/insight`, `/ranking`, `/chat` | `ia` |
| Tavily (busca web) via Worker | interno de `ia` | `ia` |
| Trace360 (dossiê assíncrono) | `POST /api/enriquecimento`, `GET /api/enriquecimento/:id` | `enriquecimento` |
| Firestore `preCadastros` | `GET/POST/PATCH/DELETE /api/precadastros` | `prospeccao` |
| Firestore `listasProspeccao` | `/api/prospeccao/listas` | `prospeccao` |
| Firestore `prospeccoes` | gravadas server-side ao buscar | `prospeccao` |
| Firestore `consultas_log` / `atividades_log` | middleware de log + `/api/admin/*` | `admin` |
| Firestore `presenca` | `POST /api/presenca` (heartbeat) | `admin` |
| Google Maps/Street View embed | permanece no frontend (iframe por endereço, sem chave secreta) | — |

> Street View/Maps continua no frontend porque é embed por endereço (não usa
> chave secreta). Se passar a usar Maps JS API com chave, ela vira `VITE_*`
> pública **só se** a chave tiver restrição de referrer; caso contrário, proxy.

---

## 5. Modelo de dados (Firestore → Postgres)

Sete coleções viram tabelas (migrations com `goose`):

| Coleção Firestore | Tabela Postgres | Observações |
|-------------------|-----------------|-------------|
| `usuarios` | `usuarios` | id, email, senha_hash (bcrypt), nome, status (pendente/aprovado), is_admin, criado_em |
| `preCadastros` | `pre_cadastros` | dados do cliente + status + notas + uid autor |
| `listasProspeccao` | `listas_prospeccao` | uid, nome, filtros, itens (jsonb) |
| `prospeccoes` | `prospeccoes` | busca registrada (uid, filtros, total, criado_em) |
| `consultas_log` | `consultas_log` | uso mensal por usuário |
| `atividades_log` | `atividades_log` | tipo, uid, nome, payload (jsonb), criado_em |
| `presenca` | `presenca` (ou Redis TTL) | quem está online — **melhor em Redis** com expiração |

**Migração dos dados vivos:** exportar do Firestore (Admin SDK/`gcloud`) e
importar via script de seed no Postgres. Ver decisão em §6.

---

## 6. Decisões que precisam de definição

Estas mudam o escopo — trago recomendação para cada uma:

1. **Firebase → Go+Postgres é obrigatório?**
   O padrão não permite Firebase. **Recomendo migrar** (auth em Go/JWT/bcrypt,
   dados em Postgres). É o maior bloco de trabalho, mas sem ele não há
   conformidade. Alternativa (fora do padrão): manter Firebase só na transição.

2. **Chart.js (dashboard admin) — não está no allowlist.**
   **Recomendo** recriar os gráficos com SVG/Tailwind à mão (barras/linhas
   simples bastam para o dashboard) ou confirmar com a infra se liberam uma lib
   de gráfico. Não usar Chart.js sem autorização.

3. **Exportação (Excel/PDF/QR) — SheetJS/html2pdf/qrcodejs não estão no allowlist.**
   **Recomendo** gerar no **backend Go**: Excel/CSV com `encoding/csv` +
   biblioteca de xlsx a validar; PDF via serviço server-side; QR via lib Go.
   Confirmar libs Go com a infra. Alternativa: CSV/JSON puro no MVP e PDF depois.

4. **Deploy: GitHub Pages sai.**
   GitHub Pages não roda backend. O alvo passa a ser Docker na **infra da
   plataforma ARCOM**. **Preciso saber onde vai rodar** (servidor/host, domínio,
   TLS) para ajustar o `docker-compose` e o `.env` de produção.

5. **Migração de dados existentes** (usuários aprovados, pré-cadastros, listas):
   fazer *cutover* único ou rodar em paralelo? **Recomendo** exportar do
   Firestore e semear no Postgres num cutover planejado, com janela curta.

6. **Segredos:** rotacionar as chaves expostas antes/durante a migração
   (Trace360, Arcom CNPJ, Geoapify) e guardá-las como env do backend.

---

## 7. Fases da migração

### Fase 0 — Preparação (sem código de app)
- Confirmar decisões da §6 com a infra/gestor.
- Rotacionar segredos expostos.
- Exportar snapshot do Firestore (backup + base do seed).

### Fase 1 — Fundação do repositório (esqueleto que sobe)
- Trazer a base da spec: `docker-compose.yml`, `.env.example`, `.gitignore`,
  `frontend/` (Vite+Tailwind+shadcn, `.npmrc`, `vite.config.ts` com proxy,
  `tailwind.config.js` com tokens ARCOM), `backend/` (Go + chi + `GET /api/health`).
- `docker compose up --build` responde em `http://localhost:8080` e
  `/api/health` retorna 200.
- **Sem** funcionalidade de negócio ainda — só a casca.

### Fase 2 — Design System + shell da UI
- Importar Red Hat Display no `src/index.css`.
- Componentes shadcn com tokens ARCOM: Button, Card, Badge, Input/Select, Alert,
  Sidebar (`bg-verde-escuro`), Topbar, layout `bg-surface`.
- Ícones `lucide-react` (substituem Font Awesome), sem emoji.
- Navegação com `react-router-dom`: rotas para as 5 abas + Login (telas vazias).

### Fase 3 — Auth (tira o Firebase do caminho)
- Backend: `internal/auth` (login, registro pendente, JWT, bcrypt, aprovação),
  tabela `usuarios`, super-admin por e-mail.
- Frontend: tela de login/cadastro, guarda de rota, `axios` com token.
- Migrar usuários do Firestore para Postgres (seed).

### Fase 4 — Consulta de CNPJ (primeiro módulo de valor)
- Backend `internal/cnpj`: proxy da API Arcom (JWT server-side) + fallback
  BrasilAPI + cache (Redis/Postgres, 24h) + `consultas_log`.
- Frontend: tela de consulta (individual/lote), validação de dígito no cliente,
  favoritos/histórico (localStorage por enquanto), "já é cliente Arcom".

### Fase 5 — Prospecção + Loja física + Pré-cadastro
- Backend `prospeccao` + `geo`: busca, corte de grandes redes, score loja física,
  ranking, `pre_cadastros`, `listas_prospeccao`, `prospeccoes`, geocode Geoapify.
- Frontend: filtros multi-cidade/ramo, tabela ordenada por score, modal
  Fachada/Mapa (Street View embed), contatos clicáveis, pré-cadastro.
- PDF de indicação (ver decisão §6.3).

### Fase 6 — IA (Insight, ranking, chat) + Enriquecimento Trace360
- Backend `ia`: Groq (insight/ranking/chat com tool-calling) + Tavily server-side
  (substitui o Cloudflare Worker). `enriquecimento`: Trace360 assíncrono
  (enfileirar com `asynq`, progresso via polling).
- Frontend: insight por empresa e em lote, ranking por IA, chat com histórico,
  dossiê Trace360 com barra de progresso.

### Fase 7 — Conversão + Admin + Exportação + acabamento
- Backend `conversao` (planilha de vendas, retorno da prospecção) e `admin`
  (dashboard, `atividades_log`, presença via Redis).
- Frontend: aba Conversão, dashboard admin (gráficos — ver §6.2), tabela de
  atividade por usuário.
- Exportação (§6.3), tema claro/escuro + cor de destaque + partículas, PWA
  (`vite-plugin-pwa` já no template), acessibilidade (`prefers-reduced-motion`).
- Cutover: apontar produção para o novo deploy, aposentar GitHub Pages/Firebase.

---

## 8. Riscos e pontos de atenção

- **Segredos já vazados** no histórico público → rotacionar (não dá para
  "desvazar"): prioridade máxima.
- **Chart.js / libs de export fora do allowlist** → definir substitutos antes de
  chegar nas fases 6/7 para não travar.
- **Migração de dados** do Firestore (usuários aprovados, pré-cadastros, listas)
  → precisa de janela de cutover e validação.
- **Paridade de features**: o app atual é rico (~250 funções). Migrar por fases
  evita "big bang", mas exige aceitar que o novo app começa com menos e cresce.
- **Deploy**: sem host definido para o backend, a Fase 1 sobe só localmente.
- **Regras do Firestore** (permissões por status/admin) viram autorização no
  backend Go — não esquecer (hoje a segurança está nas `firestore.rules`).

---

## 9. Próximo passo sugerido

Confirmadas as decisões da §6, começar pela **Fase 1** (esqueleto que sobe com
`docker compose up`), depois **Fase 3 (Auth)** — porque é o que remove o Firebase
e as maiores exposições — e seguir módulo a módulo. Cada fase entra em um
conjunto de commits pequeno e verificável na branch
`claude/arcom-ai-first-migration-qvpj17`.
