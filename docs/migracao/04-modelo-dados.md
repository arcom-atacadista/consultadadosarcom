# 04 — Modelo de dados (Firestore → Postgres + Redis)

Aplica `padroes/03` (gorm/pgx + goose) e `padroes/01` (só Postgres/Redis).
As 7 coleções do Firestore viram tabelas; o que é efêmero vai pro Redis.

## 1. De coleção para tabela

| Coleção Firestore | Destino | Motivo |
|-------------------|---------|--------|
| `usuarios` | tabela `usuarios` | persistente, relacional |
| `preCadastros` | tabela `pre_cadastros` | persistente, compartilhado |
| `listasProspeccao` | tabela `listas_prospeccao` | persistente, por usuário |
| `prospeccoes` | tabela `prospeccoes` | histórico de buscas |
| `consultas_log` | tabela `consultas_log` | uso mensal por usuário |
| `atividades_log` | tabela `atividades_log` | auditoria |
| `presenca` | **Redis** (chave com TTL) | efêmero (quem está online) |
| posse Trace360 (hoje em localStorage) | tabela `enriquecimentos` | passa a ser servidor |

## 2. Migrations (goose) — esboço das tabelas

```sql
-- 001_usuarios.sql
CREATE TABLE usuarios (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email       TEXT UNIQUE NOT NULL,
  senha_hash  TEXT NOT NULL,               -- bcrypt
  nome        TEXT NOT NULL,
  status      TEXT NOT NULL DEFAULT 'pendente', -- pendente | aprovado
  is_admin    BOOLEAN NOT NULL DEFAULT false,
  criado_em   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 002_pre_cadastros.sql
CREATE TABLE pre_cadastros (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cnpj       TEXT NOT NULL,
  razao      TEXT,
  endereco   TEXT,
  contato    TEXT,
  notas      TEXT,
  status     TEXT NOT NULL DEFAULT 'novo',
  autor_uid  UUID REFERENCES usuarios(id),
  criado_em  TIMESTAMPTZ NOT NULL DEFAULT now(),
  atualizado_em TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 003_listas_prospeccao.sql
CREATE TABLE listas_prospeccao (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  uid        UUID NOT NULL REFERENCES usuarios(id),
  nome       TEXT NOT NULL,
  filtros    JSONB NOT NULL,
  itens      JSONB NOT NULL,
  criado_em  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 004_prospeccoes.sql
CREATE TABLE prospeccoes (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  uid        UUID NOT NULL REFERENCES usuarios(id),
  filtros    JSONB NOT NULL,
  total      INTEGER NOT NULL,
  criado_em  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 005_consultas_log.sql
CREATE TABLE consultas_log (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  uid        UUID NOT NULL REFERENCES usuarios(id),
  cnpj       TEXT,
  fonte      TEXT,                          -- arcom | brasilapi
  criado_em  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_consultas_uid_mes ON consultas_log (uid, criado_em);

-- 006_atividades_log.sql
CREATE TABLE atividades_log (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tipo       TEXT NOT NULL,                 -- conta_criada|login|consulta|prospeccao|pdf|...
  uid        UUID REFERENCES usuarios(id),
  nome       TEXT,
  payload    JSONB,
  criado_em  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_atividades_tipo ON atividades_log (tipo, criado_em);

-- 007_enriquecimentos.sql
CREATE TABLE enriquecimentos (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  uid         UUID NOT NULL REFERENCES usuarios(id),
  cnpj        TEXT NOT NULL,
  cliente_id  TEXT,                         -- id do lado da Trace360
  status      TEXT NOT NULL DEFAULT 'pendente',
  dossie      JSONB,
  criado_em   TIMESTAMPTZ NOT NULL DEFAULT now(),
  atualizado_em TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (uid, cnpj)
);
```

> `gen_random_uuid()` vem da extensão `pgcrypto` (ou usar `github.com/google/uuid`
> na aplicação, do allowlist). Índices adicionais conforme necessidade de leitura
> do dashboard.

## 3. Redis (`go-redis` / `asynq`)

- **Presença:** `SET presenca:<uid> <nome> EX 60` a cada heartbeat; admin lista
  as chaves `presenca:*` vivas.
- **Cache de CNPJ:** `SET cnpj:<numero> <json> EX 86400` (24h).
- **Filas Trace360:** `asynq` para poll de status assíncrono (fase 6 do
  guia [`09`](09-fases.md)).

## 4. Migração dos dados vivos (cutover)

1. Exportar do Firestore (Admin SDK / `gcloud firestore export`) os `usuarios`
   **aprovados**, `preCadastros`, `listasProspeccao` (o histórico de logs pode
   ficar como arquivo morto — não precisa migrar).
2. Script de **seed** em Go lê o export e insere no Postgres. Senhas do Firebase
   **não migram** (hash proprietário) → usuários fazem **reset de senha** no
   primeiro acesso, ou o admin recria. Definir no guia [`10`](10-riscos-e-decisoes.md).
3. Validar contagens antes de virar o tráfego.
