# Plano de Migração — CDA (Consulta Dados Arcom) → Padrão ARCOM AI-First (v1.2)

> Documento de planejamento. Descreve **toda** a migração do app atual (um
> `index.html` monolítico em HTML/CSS/JS puro, hospedado no GitHub Pages) para o
> padrão de projetos AI-First da ARCOM (React + Vite + Tailwind/shadcn no
> frontend, NestJS **ou** FastAPI no backend, tudo sob Docker Compose).
>
> Não altera código do app ainda — serve para alinhar escopo, decisões e ordem
> de execução antes de começar.

---

## 1. Objetivo

Levar o CDA para dentro do padrão ARCOM AI-First **sem perder nenhuma
funcionalidade**, resolvendo de quebra os problemas que o padrão existe para
evitar:

- **Segredos expostos no navegador** (hoje há 3 chaves de API no código do
  frontend — qualquer usuário lê no "ver código-fonte").
- **Chamadas a domínios externos direto do browser** (o padrão exige que passem
  pelo backend).
- **Stack fora do allowlist** (Firebase, Chart.js, SheetJS, html2pdf, qrcode,
  Font Awesome não estão liberados).
- **Sem build/estrutura padrão** (arquivo único, sem `frontend/`, `backend/`,
  `docker-compose.yml`).

---

## 2. Diagnóstico do app atual

### 2.1 Formato
- `index.html` — **6.916 linhas**: ~1.500 de HTML, ~800 de CSS (já usando os
  tokens do Design System ARCOM inline), um `<script>` de **4.348 linhas** com
  toda a lógica e um segundo `<script>` de 220 linhas (ferramentas do chat de IA).
- `firestore.rules` — regras de segurança do Firestore.
- `README.md`.
- Sem build, sem `package.json`, sem pastas. Deploy no **GitHub Pages**.

### 2.2 Funcionalidades (5 abas)
| Aba | O que faz |
|---|---|
| **Consulta** | Consulta de CNPJ individual/lote, cache 24h, validação de dígito, favoritos com etiquetas, histórico, mapa/Street View, roteiro de visita, rede de sócios, **Insight IA** |
| **Prospecção** | Busca de empresas ativas por cidade/UF/CNAE, multi-cidade/multi-ramo, corte de grandes redes, score "provável loja física", listas salvas, PDF de indicação, pré-cadastro |
| **Enriquecimento** | Enriquecimento via **Trace360 AI** (fluxo em etapas, dossiê) |
| **Conversão** | Acompanhamento de conversão/vendas da prospecção (admin) |
| **Admin** | Dashboard (contas, logins, consultas, prospecções, PDFs), histórico de atividades, presença online, aprovação de usuários — com **gráficos (Chart.js)** |

### 2.3 Integrações externas e segredos (o ponto crítico)
| Serviço | URL | Chave no frontend? | Situação |
|---|---|---|---|
| Consulta CNPJ Arcom | `consultacnpj.arcom.com.br` | **`ARCOM_API_KEY` exposta** | ⛔ mover p/ backend |
| Trace360 AI | `trace360ai.arcom.com.br/api/v1` | **`TRACE360_API_KEY` exposta** | ⛔ mover p/ backend |
| Geoapify (geocode) | `api.geoapify.com` | **`GEOAPIFY_API_KEY` exposta** | ⛔ mover p/ backend |
| Groq + Tavily (IA) | via Cloudflare Worker `consultarcom...workers.dev` | não (já atrás de proxy) | ✅ padrão certo — substituir Worker pelo backend |
| BrasilAPI | `brasilapi.com.br` | pública, sem chave | mover p/ backend (regra "sem chamada externa do browser") |
| Firebase (Auth+Firestore) | SDK compat | apiKey pública (ok por design do Firebase) | ⛔ **fora do allowlist** — migrar p/ Postgres + auth do backend |
| Google Maps / Street View | `maps.google.com/...?output=embed` | — | ✅ é **iframe** (não é fetch de dados) — pode ficar |

### 2.4 Dados no Firestore (7 coleções)
`usuarios` (perfil, status pendente/aprovado, isAdmin), `consultas_log`,
`atividades_log`, `presenca` (online), `preCadastros`, `listasProspeccao`,
`prospeccoes`.

### 2.5 Bibliotecas via CDN
Firebase compat (boot); Chart.js, SheetJS/`xlsx`, html2pdf.js, qrcodejs (sob
demanda); Font Awesome; fontes Red Hat Display + JetBrains Mono.

### 2.6 Estado no navegador (localStorage — precisa continuar funcionando)
`apiSelecionada`, `arcomToken`/`arcomTokenExp`, `assessoresConhecidos`,
`chatHistorico`, `cnpjHistory`, `darkMode`, `exportColsSelecionadas`, `favTags`,
`favorites`, `navCollapsed`, `particleSettings`, `prospeccaoListasSalvas`,
`themeSettings`. (Boa parte pode continuar em localStorage; algumas idealmente
vão para o Postgres — ver §6.)

---

## 3. Gap analysis — app atual × padrão AI-First

| Área | Hoje | Padrão AI-First | Ação |
|---|---|---|---|
| Frontend | HTML/CSS/JS puro | React + Vite + TS + Tailwind + shadcn (pnpm) | Reescrever a UI |
| Gerenciador | nenhum (CDN) | **pnpm** + `.npmrc` `ignore-scripts=true` | Adotar |
| Backend | nenhum | NestJS **ou** FastAPI sob `/api`, `GET /api/health`, `0.0.0.0:3000` | **Criar** (obrigatório aqui — ver §4.1) |
| Auth/Banco | Firebase Auth + Firestore | JWT no backend + PostgreSQL | Migrar |
| Segredos | 3 chaves no bundle | só `VITE_*` públicas; segredo no backend | Mover tudo p/ backend |
| Chamadas externas | direto do browser | sempre via `/api/...` | Proxiar no backend |
| Hospedagem | GitHub Pages | Docker Compose (front :8080 + back :3000 + postgres/redis) | Trocar |
| Ícones | Font Awesome | `lucide-react` | Trocar |
| Datas | JS puro | `dayjs` | Adotar |
| Estilo | CSS inline (já ARCOM) | tokens no `tailwind.config.js` | Portar tokens |

---

## 4. Conflitos críticos e decisões necessárias

Estes pontos **mudam o plano** e precisam de decisão antes de executar.

### 4.1 Backend passa a ser obrigatório
O padrão diz "frontend primeiro, backend só quando precisar". **Aqui já
precisa**, e por vários motivos do próprio padrão: esconder segredos (3 chaves),
falar com APIs externas (Arcom, Trace360, Geoapify, Groq/Tavily, BrasilAPI),
login/contas de usuário e persistência multiusuário. Ou seja: **backend +
PostgreSQL entram desde o início.** Redis é opcional (útil p/ cache de CNPJ e
rate-limit).

**Decisão A — motor do backend:** NestJS (Node/TS) **ou** FastAPI (Python).
- Recomendação: **NestJS** — mantém tudo em TypeScript (um idioma só no
  projeto), auth JWT + TypeORM/Postgres bem mapeados, `@nestjs/axios` para
  proxiar as APIs. FastAPI é igualmente válido (mais enxuto para proxy e tem
  `jinja2` para templates).

### 4.2 Firebase → Postgres + auth do backend
Firebase (Auth e Firestore) **não está no allowlist**. Precisa migrar:
- **Auth:** e-mail/senha → `@nestjs/jwt` + `passport-jwt` + hash com
  **`bcryptjs`** (ou, no FastAPI, `pyjwt` + `passlib[bcrypt]`). Mantém o fluxo
  "novo usuário fica pendente até um admin aprovar" e o super-admin.
- **Dados:** as 7 coleções viram tabelas no Postgres (schema em §7).
- **Migração dos dados existentes:** script de export do Firestore → import no
  Postgres (usuários, pré-cadastros, listas, logs). **Decisão B:** migrar o
  histórico ou começar limpo? (Recomendo migrar usuários + pré-cadastros +
  listas; logs antigos são opcionais.)

### 4.3 Bibliotecas fora do allowlist (exports, gráficos, QR)
O allowlist do frontend é fechado. Estas libs **não** estão nele e são usadas
hoje:

| Recurso hoje | Lib atual (proibida) | Alternativa dentro do padrão |
|---|---|---|
| Exportar Excel | SheetJS `xlsx` | **CSV** feito à mão (sem lib) no front, **ou** gerar `.xlsx` **no backend** (endpoint `/api/export`) |
| Exportar PDF (indicação/dossiê) | `html2pdf.js` | Gerar no **backend** com template (`handlebars`/`jinja2`) + página pronta para impressão; download via `/api/export/pdf` |
| QR code (rota no PDF) | `qrcodejs` | Gerar no **backend** (junto do PDF) |
| Gráficos (dashboard admin) | Chart.js | Gráficos **em SVG/HTML com React + Tailwind** (barras/linhas simples, dá conta do dashboard) |
| Ícones | Font Awesome | `lucide-react` (no allowlist) |

**Decisão C — estratégia de export/gráficos:**
1. **(Recomendado)** Mover Excel/PDF/QR para o **backend** e gráficos para
   **SVG próprio** — 100% dentro do padrão, sem pedir exceção.
2. Ou **solicitar à infra a inclusão** de libs específicas no allowlist (ex.:
   uma lib de gráfico). Isso depende de aprovação da equipe que mantém o padrão.

> Pela regra 05 ("não improvise; ofereça a alternativa liberada"), a opção 1 é a
> que segue o padrão sem depender de exceção.

### 4.4 Google Maps / Street View
São **iframes** (`output=embed`), não requisições de dados do browser — podem
permanecer no frontend. O que sai do front é o **geocoding do Geoapify** (tem
chave e é chamada de dados), que vai para o backend.

---

## 5. Arquitetura alvo

```
consultadadosarcom/
├── docker-compose.yml        # front :8080 + back :3000 + postgres + redis
├── .env.example              # segredos ficam aqui (não no bundle)
├── README.md
├── frontend/                 # React + Vite + TS + Tailwind + shadcn
│   ├── .npmrc  Dockerfile  vite.config.ts  tailwind.config.js  postcss.config.js
│   ├── package.json  pnpm-lock.yaml  index.html
│   └── src/
│       ├── main.tsx  App.tsx
│       ├── lib/{api.ts, cn.ts}
│       ├── components/  (shadcn + ARCOM)
│       ├── pages/       (consulta, prospeccao, enriquecimento, conversao, admin, login)
│       └── hooks/
└── backend/                  # NestJS (ou FastAPI) — tudo sob /api
    ├── .npmrc  Dockerfile
    ├── src/ (ou app/)
    │   ├── health/                     GET /api/health
    │   ├── auth/                        login, registro pendente, JWT, guard admin
    │   ├── usuarios/                    aprovar/negar/promover
    │   ├── cnpj/                        proxy Arcom + BrasilAPI + cache
    │   ├── prospeccao/                  busca estabelecimentos + score + listas
    │   ├── enriquecimento/              proxy Trace360
    │   ├── geo/                         proxy Geoapify (geocode)
    │   ├── ia/                          proxy Groq + Tavily (substitui o Worker)
    │   ├── precadastro/                 CRUD
    │   ├── export/                      xlsx + pdf + qrcode (server-side)
    │   └── admin/                       métricas do dashboard + logs
    └── ...
```

Fluxo: navegador → `frontend :8080`; `/api/*` → proxy do Vite → `backend :3000`
→ Postgres/Redis + APIs externas. Sem CORS, sem segredo no browser.

---

## 6. Plano de migração por fases

Cada fase é entregável e testável de forma independente (`docker compose up
--build` funcionando ao fim de cada uma).

**Fase 0 — Fundação (esqueleto do padrão)**
- Criar `docker-compose.yml`, `.env.example`, `frontend/`, `backend/` a partir
  do template ARCOM. `.npmrc` obrigatório. `GET /api/health` de pé.
- Portar os tokens ARCOM do CSS atual para `frontend/tailwind.config.js`
  (paleta, Red Hat Display, raios, sombras). Base shadcn + `lib/{api,cn}`.

**Fase 1 — Autenticação (Firebase → backend/Postgres)**
- Backend: `auth/` (registro pendente, login JWT, guard de admin/aprovado),
  `usuarios/` (aprovar/negar/promover, super-admin).
- Frontend: telas de login/solicitação de acesso; guardas de rota.
- Migrar tabela `usuarios` do Firestore.

**Fase 2 — Consulta de CNPJ**
- Backend `cnpj/`: proxy Arcom (JWT interno) + BrasilAPI + lote + cache (Redis
  ou 24h) + log de consultas + cota por usuário.
- Frontend: aba Consulta (individual/lote, validação de dígito, favoritos,
  etiquetas, histórico, mapa/Street View via iframe, roteiro, rede de sócios).

**Fase 3 — Prospecção**
- Backend `prospeccao/`: busca de estabelecimentos, multi-cidade/ramo, dedupe,
  corte de grandes redes, score "provável loja física", listas salvas.
- Frontend: aba Prospecção completa + `geo/` para ordenar por proximidade.

**Fase 4 — Enriquecimento + IA**
- Backend `enriquecimento/` (proxy Trace360) e `ia/` (proxy Groq+Tavily —
  substitui o Cloudflare Worker). Frontend: abas Enriquecimento e Insight/chat.

**Fase 5 — Pré-cadastro + Exportações**
- Backend `precadastro/` (CRUD) e `export/` (xlsx + pdf + qrcode server-side).
- Frontend: fluxo de pré-cadastro, PDF de indicação, exportações.

**Fase 6 — Admin / Dashboard**
- Backend `admin/`: métricas (contas, logins, consultas, prospecções, PDFs),
  histórico de atividades, presença. Frontend: dashboard com **gráficos SVG**.

**Fase 7 — Corte e limpeza**
- Migração final de dados, remover `index.html` monolítico e `firestore.rules`,
  desligar GitHub Pages, atualizar README, validar tudo no Compose.

---

## 7. Migração de dados (Firestore → Postgres)

Esboço do schema (ajustar na execução):

- `usuarios`(id, email, senha_hash, nome, status[`pendente`|`aprovado`],
  is_admin, criado_em)
- `consultas_log`(id, usuario_id, cnpj, fonte, criado_em)
- `atividades_log`(id, usuario_id, tipo, detalhe jsonb, criado_em)
- `presenca`(usuario_id, ultimo_ping) — ou via Redis TTL
- `pre_cadastros`(id, usuario_id, cnpj, razao, dados jsonb, status, notas,
  criado_em, atualizado_em)
- `listas_prospeccao`(id, usuario_id, nome, filtros jsonb, itens jsonb, criado_em)
- `prospeccoes`(id, usuario_id, filtros jsonb, resultado jsonb, criado_em)

Processo: export do Firestore (script Node/Python) → normalização → `INSERT` via
migrations (`typeorm`/`alembic`). Usuários geram nova senha no primeiro acesso
(o hash do Firebase não é portável) — **ou** convite por e-mail para redefinir.

---

## 8. Riscos e pontos de atenção

- **Reescrita grande de UI:** 6.9k linhas → componentes React. Mitigar migrando
  aba por aba (fases 2–6), com o app subindo a cada fase.
- **Paridade de features:** score de loja física, corte de redes, cache e cota
  têm lógica de negócio embutida no HTML — portar com cuidado e testes.
- **Senhas do Firebase não migram:** exige redefinição no primeiro login.
- **Exports/gráficos:** decisão §4.3 muda esforço (server-side vs. exceção no
  allowlist).
- **Rate limits das APIs** (Groq/Tavily/Arcom): centralizar no backend ajuda a
  controlar e cachear.

---

## 9. Decisões pendentes (para destravar a execução)

- **A.** Motor do backend: **NestJS** (recomendado) ou FastAPI?
- **B.** Migrar histórico do Firestore (usuários/pré-cadastros/listas/logs) ou
  começar com base limpa?
- **C.** Export/gráficos: mover para o **backend + SVG** (dentro do padrão) ou
  pedir exceção de allowlist à infra?
- **D.** Redis desde já (cache de CNPJ + rate-limit) ou só Postgres no começo?

Confirmadas essas quatro, começo pela **Fase 0** (esqueleto do padrão) e sigo as
fases na ordem acima.
