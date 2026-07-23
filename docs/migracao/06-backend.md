# 06 — Backend (Go)

Aplica `padroes/03` e `padroes/01`. Rotas sob `/api`, `GET /api/health`, escuta
em `0.0.0.0:3000`, build estático `CGO_ENABLED=0`, segredos só via ambiente.

## 1. Pacotes (`internal/`)

```
backend/
├── cmd/server/main.go        # monta o router chi, sobe na :3000
└── internal/
    ├── config/               # lê env (DATABASE_URL, REDIS_URL, PORT, chaves)
    ├── db/                   # conexão Postgres (gorm/pgx) + goose migrations
    ├── redis/                # cliente go-redis + asynq
    ├── http/                 # router, middlewares, health, erros padronizados
    ├── auth/                 # login, registro, JWT, bcrypt, aprovação
    ├── usuarios/             # CRUD/admin de contas
    ├── cnpj/                 # token Arcom + /v1/cnpj/batch + BrasilAPI + cache
    ├── prospeccao/           # /v1/municipios, /v1/estabelecimentos, score, listas, precadastros
    ├── geo/                  # proxy Geoapify
    ├── enriquecimento/       # Trace360 (POST /clientes + status), fila asynq
    ├── ia/                   # Groq (insight/ranking/chat) + Tavily (tools)
    ├── conversao/            # planilha de vendas + retorno da prospecção
    ├── admin/                # dashboard, atividades, presença
    └── atividades/           # middleware de log (consultas_log, atividades_log)
```

Um pacote por recurso (padroes/03/06): cada um com `model` (struct/tabela),
`repo` (acesso a dados), `service` (regra) e `handler` (rota chi).

## 2. Ponto de entrada (padroes/03)

```go
r := chi.NewRouter()
r.Use(middleware.Logger, middleware.Recoverer)
r.Route("/api", func(api chi.Router) {
    api.Get("/health", health)                 // público
    api.Mount("/auth", authRouter)             // público (login/register)
    api.Group(func(pr chi.Router) {            // exige JWT + aprovado
        pr.Use(RequireAuth, RequireAprovado, LogAtividade)
        pr.Mount("/cnpj", cnpjRouter)
        pr.Mount("/prospeccao", prospeccaoRouter)
        pr.Mount("/precadastros", precadastroRouter)
        pr.Mount("/geo", geoRouter)
        pr.Mount("/enriquecimento", enriqRouter)
        pr.Mount("/ia", iaRouter)
        pr.Mount("/conversao", conversaoRouter)
        pr.Post("/presenca", presencaHandler)
        pr.Group(func(ar chi.Router) {         // exige admin
            ar.Use(RequireAdmin)
            ar.Mount("/usuarios", usuariosRouter)
            ar.Mount("/admin", adminRouter)
        })
    })
})
http.ListenAndServe("0.0.0.0:"+port, r)
```

## 3. Responsabilidades por pacote

- **config:** `os.Getenv` (+ `godotenv` em dev). Falha cedo se faltar chave
  obrigatória. Nunca hardcode.
- **auth:** senha com `bcrypt`; token com `golang-jwt/jwt/v5` (claims: uid,
  email, isAdmin, status). Registro sempre `pendente`/`isAdmin=false`.
  Super-admin por `SUPER_ADMIN_EMAIL`.
- **cnpj:** guarda o token da API Arcom em memória/Redis (troca via
  `POST /v1/auth/token`), reautentica em 401, chunk de 1000, normaliza o formato
  aninhado da Arcom e o formato BrasilAPI num DTO único; cache 24h em Redis;
  valida dígito verificador antes de chamar terceiro.
- **prospeccao:** pagina `/v1/estabelecimentos`, dedup, **corte de grandes
  redes**, **score de loja física** (regra do guia [`01`](01-diagnostico-atual.md) §6),
  listas e pré-cadastros.
- **geo:** proxy Geoapify (chave server-side).
- **enriquecimento:** `POST /clientes` na Trace360 com `x-api-key` server-side;
  grava posse em `enriquecimentos`; `asynq`/`cron` atualiza status.
- **ia:** chama Groq/Tavily server-side; implementa as ferramentas
  `buscar_web` e `estatisticas_do_site` (esta lê agregados do backend,
  respeitando permissão).
- **conversao:** recebe planilha (multipart), parseia (ver [`08`](08-conformidade-allowlist.md)),
  cruza com prospecções.
- **admin/atividades:** agregações do dashboard, histórico filtrável, presença
  (Redis), e o **middleware** que grava `consultas_log`/`atividades_log` com o
  usuário do JWT.

## 4. Módulos Go (todos do allowlist — padroes/01)

`go-chi/chi/v5`, `log/slog`, `joho/godotenv`, `gorm`+`driver/postgres` (ou
`jackc/pgx/v5`), `pressly/goose/v3`, `redis/go-redis/v9`, `hibiken/asynq`,
`x/crypto/bcrypt`, `golang-jwt/jwt/v5`, `go-playground/validator/v10`,
`robfig/cron/v3`, `google/uuid`, stdlib `net/http`/`encoding/json`/`encoding/csv`/
`html/template`/`time`. **Nada fora disso.**

## 5. Segurança (padroes/03)

- Validar **tudo** que vem do cliente (`validator`). Nunca confiar no front.
- Commite `go.sum` (integridade dos módulos).
- Segredos só via ambiente; nunca no código nem no repo.
- Autorização por middleware (RequireAuth/Aprovado/Admin) — ver [`07`](07-seguranca.md).
