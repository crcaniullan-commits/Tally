-- Metodo de pago en los gastos, igual que en los ingresos (migrations/000001).
--
-- Reutiliza el ENUM payment_method que ya existe: 'debito', 'credito',
-- 'transferencia', 'efectivo'.
--
-- La columna queda NOT NULL, como en incomes. Las filas de gastos que ya
-- existen no tienen metodo de pago guardado, asi que se rellenan con
-- 'efectivo' (el metodo por defecto) para que el NOT NULL no rompa la
-- migracion. Despues se saca el DEFAULT: la API siempre manda el metodo de
-- pago, asi que no tiene que quedar un valor implicito en la base.

ALTER TABLE expenses
ADD COLUMN payment_method payment_method NOT NULL DEFAULT 'efectivo';

ALTER TABLE expenses
ALTER COLUMN payment_method DROP DEFAULT;
