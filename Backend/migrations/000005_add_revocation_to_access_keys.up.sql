-- Revocacion de una llave de acceso, ortogonal al canje (redeemed_at) y al
-- vencimiento (expires_at) que ya existen: canjear es "alguien la uso para
-- registrarse", revocar es "esta llave ya no sirve mas".
--
-- Ambas columnas quedan NULLABLES a proposito: la mayoria de las llaves nunca se
-- revocan, y revoked_at NULL es justamente el estado "llave activa".
-- revoked_by guarda quien revoco, para poder auditarlo.

ALTER TABLE access_keys
ADD COLUMN revoked_at TIMESTAMPTZ;

ALTER TABLE access_keys
ADD COLUMN revoked_by UUID REFERENCES users(id);
