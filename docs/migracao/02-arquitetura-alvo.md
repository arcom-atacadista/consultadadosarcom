# 02 — Arquitetura alvo

Aplica `padroes/00`, `02`, `03`, `04` e `06` ao CDA.

## 1. Estrutura do repositório

```
consultadadosarcom/
├── docker-compose.yml          # sobe front + back + postgres + redis
├── .env.example                # PROJECT_NAME, POSTGRES_PASSWORD, chaves de API
├── README.md                   # como rodar (curto)
├── firestore.rules             # mantido só como referência histórica
├── index.html                  # app antigo, mantido como referência de paridade
├── docs/migracao/              # esta documentação
│
├── frontend/                   # React + Vite (só desenha e chama /api)
│   ├── .npmrc                  # NÃO alterar (ignore-scripts=true)
│   ├── Dockerfile              # build + vite preview
│   ├── vite.config.ts          # proxy /api + PWA (não mexer no proxy)
│   ├── tailwind.config.js      # tokens do Design System ARCOM
│   ├── postcss.config.js
│   ├── package.json
│   ├── pnpm-lock.yaml          # obrigatório (build usa --frozen-lockfile)
│   ├── index.html
│   └── src/
│       ├── main.tsx            # entrada + router
│       ├── App.tsx
│       ├── index.css           # @import Red Hat Display + @tailwind
│       ├── lib/{api.ts,cn.ts}
│       ├── components/ui/      # shadcn com tokens ARCOM
│       ├── components/         # componentes de domínio
│       ├── pages/              # 1 por rota (ver guia 05)
│       └── hooks/
│
└── backend/                    # Go (lógica, segredos, dados)
    ├── Dockerfile              # multi-stage estático (golang:1.23 → alpine:3.20)
    ├── go.mod
    ├── go.sum                  # obrigatório
    ├── .env.example
    ├── cmd/server/main.go      # router chi + GET /api/health
    └── internal/               # ver guia 06
```

> `index.html` e `firestore.rules` **ficam na raiz** durante a transição, como
> referência de paridade e das regras de autorização. Removidos só no cutover
> final (fase 7 do guia [`09`](09-fases.md)).

## 2. Fluxo de rede (sem Traefik, sem CORS)

```
navegador ─▶ frontend (vite preview :4173 → publicado em :8080)
                 │  /api/*  ──proxy──▶ backend (chi :3000) ─┬─▶ postgres:5432
                 │                                          ├─▶ redis:6379
                 └  resto = a própria SPA                   └─▶ APIs externas:
                                                               Arcom CNPJ, BrasilAPI,
                                                               Trace360, Geoapify,
                                                               Groq, Tavily
```

- Só o **frontend** publica porta pro host (`8080:4173`).
- Backend, postgres e redis só na rede interna do compose.
- Front alcança o backend por `backend:3000` (proxy do `vite.config.ts`).
- **Toda** chamada externa que hoje sai do navegador passa a sair do backend Go.

## 3. docker-compose (base da spec, adaptada)

Usa a base de `arcom-projeto/docker-compose.yml` **sem alterar a infra**. O CDA
precisa de Postgres (dados) e Redis (presença/cache/fila Trace360), então
**mantém os quatro serviços**. As chaves de API externas entram como ambiente do
serviço `backend` (nunca no frontend):

```yaml
# trecho do serviço backend (além de DATABASE_URL/REDIS_URL/PORT já existentes)
    environment:
      ARCOM_API_BASE_URL: "https://consultacnpj.arcom.com.br"
      ARCOM_API_KEY:      "${ARCOM_API_KEY}"
      TRACE360_BASE:      "https://trace360ai.arcom.com.br/api/v1"
      TRACE360_API_KEY:   "${TRACE360_API_KEY}"
      GEOAPIFY_API_KEY:   "${GEOAPIFY_API_KEY}"
      GROQ_API_KEY:       "${GROQ_API_KEY}"
      TAVILY_API_KEY:     "${TAVILY_API_KEY}"
      JWT_SECRET:         "${JWT_SECRET}"
      SUPER_ADMIN_EMAIL:  "${SUPER_ADMIN_EMAIL}"
```

Os valores vêm de um `.env` na raiz (não commitado) ou da infra em produção.
`.env.example` lista as chaves **sem valores reais**.

## 4. Como sobe

```bash
docker compose up --build     # abre em http://localhost:8080
```

O healthcheck do backend bate em `GET /api/health` (já previsto no compose da
spec). Frontend depende do backend; backend depende de postgres/redis saudáveis.

## 5. O que se pode e não se pode mexer (padroes/00 e 04)

- **Mexe:** código em `frontend/src` e `backend/internal`, dependências do
  allowlist, bloco `manifest` do PWA no `vite.config.ts`, variáveis de ambiente.
- **Não mexe:** `docker-compose.yml` estrutural, `Dockerfile`, `.npmrc`, o proxy
  do `vite.config.ts`, os serviços de infra. (Adicionar variáveis de ambiente do
  backend é permitido; alterar a topologia dos serviços, não.)
