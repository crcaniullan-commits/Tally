-- Referencia a categoria en ingresos. Reutiliza la tabla categories, que hasta
-- ahora solo categorizaba gastos (ver migrations/000001).
--
-- Queda NULLABLE a proposito, a diferencia de expenses.category_id que es NOT
-- NULL: incomes puede tener filas guardadas y no hay categorias sembradas de
-- las que sacar un valor, asi que un NOT NULL obligaria a sembrar categorias y
-- despues hacer UPDATE sobre las filas existentes.
--
-- Ojo: el esquema no tiene ningun indice. Si el listado de ingresos termina
-- filtrando por categoria, el indice se agrega en su propia migracion.
--
-- Pendiente de cablear: el filtro por categoria todavia no existe en el store
-- (GetAllIncomesOfUser solo acota por user_id y por rango de fechas).

ALTER TABLE incomes
ADD COLUMN category_id UUID REFERENCES categories(id);
