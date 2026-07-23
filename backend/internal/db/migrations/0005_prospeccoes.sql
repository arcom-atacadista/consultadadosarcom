-- +goose Up
CREATE TABLE prospeccoes (
  id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  uid       UUID NOT NULL REFERENCES usuarios(id),
  filtros   JSONB NOT NULL,
  total     INTEGER NOT NULL,
  criado_em TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE prospeccoes;
