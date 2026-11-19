CREATE TABLE device_tokens (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token        TEXT NOT NULL UNIQUE,
    platform     TEXT NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_device_tokens_user ON device_tokens (user_id);

CREATE TABLE notification_log (
    user_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind     TEXT NOT NULL,
    ref_date DATE NOT NULL,
    sent_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, kind, ref_date)
);