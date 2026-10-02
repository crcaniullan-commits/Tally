-- Los montos pasan de NUMERIC(12,2) a BIGINT.
-- CLP tiene exponente 0 (no usa centavos), asi que el peso es la unidad minima
-- y se representa exacto como entero. El cast es explicito porque Postgres no
-- hace conversion implicita de numeric a bigint.
--
-- AVISO: ::bigint redondea (half away from zero). Si hay montos con decimales
-- guardados, 1500.50 se convierte en 1501. Verificar antes de aplicar:
--   SELECT 'incomes' t, count(*) FROM incomes WHERE monto <> trunc(monto)
--   UNION ALL SELECT 'expenses', count(*) FROM expenses WHERE monto <> trunc(monto)
--   UNION ALL SELECT 'debtors', count(*) FROM debtors WHERE monto <> trunc(monto)
--   UNION ALL SELECT 'goals', count(*) FROM goals WHERE monto_meta <> trunc(monto_meta);

ALTER TABLE incomes
ALTER COLUMN monto TYPE BIGINT USING monto::bigint;

ALTER TABLE expenses
ALTER COLUMN monto TYPE BIGINT USING monto::bigint;

ALTER TABLE debtors
ALTER COLUMN monto TYPE BIGINT USING monto::bigint;

ALTER TABLE goals
ALTER COLUMN monto_meta TYPE BIGINT USING monto_meta::bigint;
