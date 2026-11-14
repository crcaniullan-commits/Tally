-- Archivos .xlsx subidos por el usuario para una empresa. El binario vive en
-- el storage (disco por ahora, S3 despues) y en la base queda solo el
-- registro: id, nombre original, clave del objeto y tamaño.
--
-- La tenencia es por empresa: al borrar la empresa se borran sus archivos.

CREATE TABLE excel_files (
    id          UUID PRIMARY KEY,
    empresa_id  UUID NOT NULL REFERENCES empresas(id) ON DELETE CASCADE,
    nombre      TEXT NOT NULL,
    storage_key TEXT NOT NULL UNIQUE,
    size_bytes  BIGINT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_excel_files_empresa ON excel_files (empresa_id, created_at DESC);
