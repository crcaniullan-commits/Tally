-- Codigos de recuperacion de contrasena.
--
-- Un codigo de un solo uso, de 6 digitos, que se manda por correo para
-- verificar al usuario y permitirle cambiar la contrasena.
--
-- La columna `codigo` guarda el hash SHA-256 del codigo real: en la base nunca
-- se escribe el codigo que el usuario recibio, asi que un dump de la base no
-- sirve para hacerse pasar por nadie.
--
-- `verificado_en` marca que el usuario ya introdujo el codigo correctamente;
-- `usado_en` que el codigo se consumio (al cambiar la contrasena o al pedir
-- uno nuevo). Los dos son NULL en el estado inicial, que es el estado "codigo
-- vigente sin usar".

CREATE TABLE password_reset_codes (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    codigo        VARCHAR(64) NOT NULL,
    expira_en     TIMESTAMPTZ NOT NULL,
    verificado_en TIMESTAMPTZ,
    usado_en      TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_password_reset_codes_user_id ON password_reset_codes(user_id);
