CREATE TABLE excel_files (
    id          UUID PRIMARY KEY,
    empresa_id  UUID NOT NULL REFERENCES empresas(id) ON DELETE CASCADE,
    nombre      TEXT NOT NULL,
    storage_key TEXT NOT NULL UNIQUE,
    size_bytes  BIGINT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_excel_files_empresa ON excel_files (empresa_id, created_at DESC);