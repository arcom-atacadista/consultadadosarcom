-- +goose Up
CREATE TABLE pre_cadastros (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cnpj          TEXT NOT NULL,
  razao         TEXT,
  endereco      TEXT,
  contato       TEXT,
  notas         TEXT,
  status        TEXT NOT NULL DEFAULT 'novo',
  autor_uid     UUID REFERENCES usuarios(id),
  criado_em     TIMESTAMPTZ NOT NULL DEFAULT now(),
  atualizado_em TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE pre_cadastros;
