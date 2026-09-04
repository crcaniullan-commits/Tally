-- Tipos enumerados
CREATE TYPE user_role AS ENUM ('usuario', 'municipal');
CREATE TYPE payment_method AS ENUM ('debito', 'credito', 'transferencia', 'efectivo');
CREATE TYPE goal_period AS ENUM ('dia', 'semana', 'mes');

-- Usuarios
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    nombre        VARCHAR(150) NOT NULL,
    role          user_role NOT NULL DEFAULT 'usuario',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Claves de acceso a cuentas de pago, generadas por usuarios municipales
CREATE TABLE access_keys (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code         VARCHAR(50) UNIQUE NOT NULL,
    created_by   UUID NOT NULL REFERENCES users(id),
    redeemed_by  UUID REFERENCES users(id),
    redeemed_at  TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Categorías de gastos (predefinidas si user_id es null, personalizadas si no)
CREATE TABLE categories (
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre  VARCHAR(100) NOT NULL,
    user_id UUID REFERENCES users(id)
);

-- Ingresos
CREATE TABLE incomes (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id),
    monto          NUMERIC(12,2) NOT NULL,
    payment_method payment_method NOT NULL,
    descripcion    TEXT,
    fecha          DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Gastos
CREATE TABLE expenses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    monto       NUMERIC(12,2) NOT NULL,
    category_id UUID NOT NULL REFERENCES categories(id),
    descripcion TEXT,
    fecha       DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Deudores pendientes
CREATE TABLE debtors (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id),
    nombre       VARCHAR(150) NOT NULL,
    monto        NUMERIC(12,2) NOT NULL,
    descripcion  TEXT,
    pagado       BOOLEAN NOT NULL DEFAULT false,
    fecha_limite DATE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Metas mensuales de ingreso
CREATE TABLE goals (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id),
    nombre        VARCHAR(150) NOT NULL,
    periodo       goal_period NOT NULL,
    monto_meta    NUMERIC(12,2) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);