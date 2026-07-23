-- +goose Up
CREATE TABLE consultas_log (
  id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  uid       UUID NOT NULL REFERENCES usuarios(id),
  cnpj      TEXT NOT NULL,
  fonte     TEXT NOT NULL,
  criado_em TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_consultas_uid_mes ON consultas_log (uid, criado_em);

-- +goose Down
DROP TABLE consultas_log;
