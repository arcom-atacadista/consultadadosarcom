-- +goose Up
CREATE TABLE enriquecimentos (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  uid           UUID NOT NULL REFERENCES usuarios(id),
  cnpj          TEXT NOT NULL,
  cliente_id    TEXT,
  status        TEXT NOT NULL DEFAULT 'pendente',
  razao_social  TEXT,
  criado_em     TIMESTAMPTZ NOT NULL DEFAULT now(),
  atualizado_em TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (uid, cnpj)
);

-- +goose Down
DROP TABLE enriquecimentos;
