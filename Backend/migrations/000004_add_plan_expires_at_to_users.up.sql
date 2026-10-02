-- Vencimiento del plan del usuario. NULL = sin plan pago (plan gratuito), asi
-- que el filtro de quien tiene plan vigente es plan_expires_at > now(): con NULL
-- esa comparacion da NULL, o sea falso, que es justo lo que queremos.
--
-- No confundir con access_keys.expires_at, que es el vencimiento de una llave
-- de acceso puntual.

ALTER TABLE users
ADD COLUMN plan_expires_at TIMESTAMPTZ;
