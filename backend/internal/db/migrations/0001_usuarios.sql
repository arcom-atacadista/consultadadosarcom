-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE usuarios (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email      TEXT UNIQUE NOT NULL,
  senha_hash TEXT NOT NULL,
  nome       TEXT NOT NULL,
  status     TEXT NOT NULL DEFAULT 'pendente',
  is_admin   BOOLEAN NOT NULL DEFAULT false,
  criado_em  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE usuarios;
