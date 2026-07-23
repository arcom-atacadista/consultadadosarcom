-- +goose Up
CREATE TABLE listas_prospeccao (
  id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  uid       UUID NOT NULL REFERENCES usuarios(id),
  nome      TEXT NOT NULL,
  filtros   JSONB NOT NULL,
  itens     JSONB NOT NULL,
  criado_em TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE listas_prospeccao;
