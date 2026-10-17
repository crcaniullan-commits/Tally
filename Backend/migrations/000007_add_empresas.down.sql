-- Reconstruye user_id desde empresas.user_id. La relacion empresa->usuario es
-- 1:N, asi que cada fila recupera su usuario original. Si un usuario llego a
-- tener varias empresas, sus datos vuelven a quedar todos bajo el mismo user_id
-- y la distincion entre empresas se pierde de forma irrecuperable.

-- Revierte las FK al comportamiento por defecto (sin cascada ni SET NULL) y
-- devuelve created_by a NOT NULL. AVISO: si algun municipal fue borrado, su
-- created_by quedo en NULL y el SET NOT NULL aborta la migracion; ese dato no
-- se puede reconstruir. Las FK de empresa_id se caen solas al dropear la
-- columna/la tabla, asi que no hace falta tocarlas.
ALTER TABLE access_keys DROP CONSTRAINT access_keys_created_by_fkey;
ALTER TABLE access_keys ADD CONSTRAINT access_keys_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id);
ALTER TABLE access_keys ALTER COLUMN created_by SET NOT NULL;

ALTER TABLE access_keys DROP CONSTRAINT access_keys_redeemed_by_fkey;
ALTER TABLE access_keys ADD CONSTRAINT access_keys_redeemed_by_fkey
    FOREIGN KEY (redeemed_by) REFERENCES users(id);

ALTER TABLE access_keys DROP CONSTRAINT access_keys_revoked_by_fkey;
ALTER TABLE access_keys ADD CONSTRAINT access_keys_revoked_by_fkey
    FOREIGN KEY (revoked_by) REFERENCES users(id);

ALTER TABLE incomes DROP CONSTRAINT incomes_category_id_fkey;
ALTER TABLE incomes ADD CONSTRAINT incomes_category_id_fkey
    FOREIGN KEY (category_id) REFERENCES categories(id);

ALTER TABLE expenses DROP CONSTRAINT expenses_category_id_fkey;
ALTER TABLE expenses ADD CONSTRAINT expenses_category_id_fkey
    FOREIGN KEY (category_id) REFERENCES categories(id);

-- categories
ALTER TABLE categories ADD COLUMN user_id UUID REFERENCES users(id);

UPDATE categories c
SET user_id = e.user_id
FROM empresas e
WHERE e.id = c.empresa_id;

ALTER TABLE categories DROP COLUMN empresa_id;

-- goals
ALTER TABLE goals ADD COLUMN user_id UUID REFERENCES users(id);

UPDATE goals g
SET user_id = e.user_id
FROM empresas e
WHERE e.id = g.empresa_id;

ALTER TABLE goals ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE goals DROP COLUMN empresa_id;

-- debtors
ALTER TABLE debtors ADD COLUMN user_id UUID REFERENCES users(id);

UPDATE debtors d
SET user_id = e.user_id
FROM empresas e
WHERE e.id = d.empresa_id;

ALTER TABLE debtors ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE debtors DROP COLUMN empresa_id;

-- expenses
ALTER TABLE expenses ADD COLUMN user_id UUID REFERENCES users(id);

UPDATE expenses ex
SET user_id = e.user_id
FROM empresas e
WHERE e.id = ex.empresa_id;

ALTER TABLE expenses ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE expenses DROP COLUMN empresa_id;

-- incomes
ALTER TABLE incomes ADD COLUMN user_id UUID REFERENCES users(id);

UPDATE incomes i
SET user_id = e.user_id
FROM empresas e
WHERE e.id = i.empresa_id;

ALTER TABLE incomes ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE incomes DROP COLUMN empresa_id;

DROP TABLE empresas;
