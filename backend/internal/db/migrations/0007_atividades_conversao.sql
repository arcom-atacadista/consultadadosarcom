-- +goose Up
CREATE TABLE atividades_log (
  id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tipo      TEXT NOT NULL,
  uid       UUID REFERENCES usuarios(id),
  nome      TEXT,
  email     TEXT,
  detalhe   TEXT,
  payload   JSONB,
  criado_em TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_atividades_tipo ON atividades_log (tipo, criado_em);
CREATE INDEX idx_atividades_criado ON atividades_log (criado_em DESC);

ALTER TABLE listas_prospeccao ADD COLUMN assessor TEXT NOT NULL DEFAULT '';
ALTER TABLE listas_prospeccao ADD COLUMN nome_usuario TEXT NOT NULL DEFAULT '';
ALTER TABLE listas_prospeccao ADD COLUMN email TEXT NOT NULL DEFAULT '';
ALTER TABLE listas_prospeccao ADD COLUMN convertidos INTEGER;
ALTER TABLE listas_prospeccao ADD COLUMN total_empresas INTEGER;
ALTER TABLE listas_prospeccao ADD COLUMN verificado_em TIMESTAMPTZ;

-- +goose Down
ALTER TABLE listas_prospeccao DROP COLUMN verificado_em;
ALTER TABLE listas_prospeccao DROP COLUMN total_empresas;
ALTER TABLE listas_prospeccao DROP COLUMN convertidos;
ALTER TABLE listas_prospeccao DROP COLUMN email;
ALTER TABLE listas_prospeccao DROP COLUMN nome_usuario;
ALTER TABLE listas_prospeccao DROP COLUMN assessor;
DROP TABLE atividades_log;
