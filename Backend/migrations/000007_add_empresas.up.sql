-- Las empresas pasan a ser el eje de la tenencia de datos: un usuario puede
-- tener varias empresas y cada empresa pertenece a un solo usuario. incomes,
-- expenses, debtors, goals y categories dejan de referenciar users y pasan a
-- referenciar empresas.
--
-- El backfill crea una empresa por cada usuario existente y adopta users.nombre
-- como nombre_empresa: el esquema no guardaba el nombre de ningun negocio, asi
-- que no hay un valor mejor que reconstruir. Un usuario queda con una sola
-- empresa aunque despues pueda crear mas.

CREATE TABLE empresas (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre_empresa VARCHAR(150) NOT NULL,
    user_id        UUID NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_empresas_user_id ON empresas(user_id);

INSERT INTO empresas (nombre_empresa, user_id)
SELECT nombre, id FROM users;

-- incomes
ALTER TABLE incomes ADD COLUMN empresa_id UUID REFERENCES empresas(id);

UPDATE incomes i
SET empresa_id = e.id
FROM empresas e
WHERE e.user_id = i.user_id;

ALTER TABLE incomes ALTER COLUMN empresa_id SET NOT NULL;
ALTER TABLE incomes DROP COLUMN user_id;
CREATE INDEX idx_incomes_empresa_id ON incomes(empresa_id);

-- expenses
ALTER TABLE expenses ADD COLUMN empresa_id UUID REFERENCES empresas(id);

UPDATE expenses ex
SET empresa_id = e.id
FROM empresas e
WHERE e.user_id = ex.user_id;

ALTER TABLE expenses ALTER COLUMN empresa_id SET NOT NULL;
ALTER TABLE expenses DROP COLUMN user_id;
CREATE INDEX idx_expenses_empresa_id ON expenses(empresa_id);

-- debtors
ALTER TABLE debtors ADD COLUMN empresa_id UUID REFERENCES empresas(id);

UPDATE debtors d
SET empresa_id = e.id
FROM empresas e
WHERE e.user_id = d.user_id;

ALTER TABLE debtors ALTER COLUMN empresa_id SET NOT NULL;
ALTER TABLE debtors DROP COLUMN user_id;
CREATE INDEX idx_debtors_empresa_id ON debtors(empresa_id);

-- goals
ALTER TABLE goals ADD COLUMN empresa_id UUID REFERENCES empresas(id);

UPDATE goals g
SET empresa_id = e.id
FROM empresas e
WHERE e.user_id = g.user_id;

ALTER TABLE goals ALTER COLUMN empresa_id SET NOT NULL;
ALTER TABLE goals DROP COLUMN user_id;
CREATE INDEX idx_goals_empresa_id ON goals(empresa_id);

-- categories: nullable a proposito, porque las categorias globales (user_id IS
-- NULL) no pertenecen a ninguna empresa y se conservan como empresa_id NULL.
ALTER TABLE categories ADD COLUMN empresa_id UUID REFERENCES empresas(id);

UPDATE categories c
SET empresa_id = e.id
FROM empresas e
WHERE e.user_id = c.user_id;

ALTER TABLE categories DROP COLUMN user_id;
CREATE INDEX idx_categories_empresa_id ON categories(empresa_id);

-- Borrado en cascada: al eliminar un usuario se borran sus empresas y, con
-- ellas, los datos que cuelgan (incomes, expenses, debtors, goals, categories).
-- Las FK se declararon arriba sin clausula, asi que hay que recrearlas.
ALTER TABLE empresas DROP CONSTRAINT empresas_user_id_fkey;
ALTER TABLE empresas ADD CONSTRAINT empresas_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE incomes DROP CONSTRAINT incomes_empresa_id_fkey;
ALTER TABLE incomes ADD CONSTRAINT incomes_empresa_id_fkey
    FOREIGN KEY (empresa_id) REFERENCES empresas(id) ON DELETE CASCADE;

ALTER TABLE expenses DROP CONSTRAINT expenses_empresa_id_fkey;
ALTER TABLE expenses ADD CONSTRAINT expenses_empresa_id_fkey
    FOREIGN KEY (empresa_id) REFERENCES empresas(id) ON DELETE CASCADE;

ALTER TABLE debtors DROP CONSTRAINT debtors_empresa_id_fkey;
ALTER TABLE debtors ADD CONSTRAINT debtors_empresa_id_fkey
    FOREIGN KEY (empresa_id) REFERENCES empresas(id) ON DELETE CASCADE;

ALTER TABLE goals DROP CONSTRAINT goals_empresa_id_fkey;
ALTER TABLE goals ADD CONSTRAINT goals_empresa_id_fkey
    FOREIGN KEY (empresa_id) REFERENCES empresas(id) ON DELETE CASCADE;

ALTER TABLE categories DROP CONSTRAINT categories_empresa_id_fkey;
ALTER TABLE categories ADD CONSTRAINT categories_empresa_id_fkey
    FOREIGN KEY (empresa_id) REFERENCES empresas(id) ON DELETE CASCADE;

-- Al borrar una categoria se limpia la referencia del ingreso (la columna es
-- nullable) y se borra el gasto (su category_id es NOT NULL, no admite NULL).
-- Sin esto, la cascada de empresa podria borrar la categoria antes que las
-- filas que la referencian y reventar por FK.
ALTER TABLE incomes DROP CONSTRAINT incomes_category_id_fkey;
ALTER TABLE incomes ADD CONSTRAINT incomes_category_id_fkey
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;

ALTER TABLE expenses DROP CONSTRAINT expenses_category_id_fkey;
ALTER TABLE expenses ADD CONSTRAINT expenses_category_id_fkey
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE;

-- access_keys NO se borra con el usuario: se pierde la autoria/canje/revocacion
-- (las columnas quedan NULL). created_by estaba NOT NULL, asi que hay que
-- soltarlo para que el SET NULL sea posible.
ALTER TABLE access_keys ALTER COLUMN created_by DROP NOT NULL;

ALTER TABLE access_keys DROP CONSTRAINT access_keys_created_by_fkey;
ALTER TABLE access_keys ADD CONSTRAINT access_keys_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE access_keys DROP CONSTRAINT access_keys_redeemed_by_fkey;
ALTER TABLE access_keys ADD CONSTRAINT access_keys_redeemed_by_fkey
    FOREIGN KEY (redeemed_by) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE access_keys DROP CONSTRAINT access_keys_revoked_by_fkey;
ALTER TABLE access_keys ADD CONSTRAINT access_keys_revoked_by_fkey
    FOREIGN KEY (revoked_by) REFERENCES users(id) ON DELETE SET NULL;
