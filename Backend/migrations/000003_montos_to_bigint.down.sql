-- Revierte a NUMERIC(12,2). Los centavos se pierden de forma irrecuperable:
-- un monto que era 1500 queda en 1500.00 y no hay forma de reconstruir el
-- estado anterior. Usar solo en desarrollo o recien migrado.
ALTER TABLE incomes
ALTER COLUMN monto TYPE NUMERIC(12,2) USING monto::numeric(12,2);

ALTER TABLE expenses
ALTER COLUMN monto TYPE NUMERIC(12,2) USING monto::numeric(12,2);

ALTER TABLE debtors
ALTER COLUMN monto TYPE NUMERIC(12,2) USING monto::numeric(12,2);

ALTER TABLE goals
ALTER COLUMN monto_meta TYPE NUMERIC(12,2) USING monto_meta::numeric(12,2);
