-- Tipo de la categoría: una categoría es de ingresos o de gastos, nunca de ambos.
CREATE TYPE category_type AS ENUM ('ingreso', 'gasto');

ALTER TABLE categories
    ADD COLUMN tipo category_type NOT NULL DEFAULT 'ingreso';

-- Las categorías que ya tienen gastos asociados son de tipo 'gasto'.
-- Los ingresos solo referencian categorías de tipo 'ingreso' y los gastos solo
-- de tipo 'gasto', así que una categoría con gastos no puede quedar como
-- 'ingreso'.
UPDATE categories AS c
SET tipo = 'gasto'
WHERE EXISTS (
    SELECT 1
    FROM expenses e
    WHERE e.category_id = c.id
);

-- Las categorías globales (Ventas, Servicios, Pedidos) son de ventas.
UPDATE categories
SET tipo = 'ingreso'
WHERE empresa_id IS NULL;
