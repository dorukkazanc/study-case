CREATE TABLE IF NOT EXISTS batches (
    id          UUID PRIMARY KEY,
    total       INT NOT NULL DEFAULT 0,
    pending     INT NOT NULL DEFAULT 0,
    queued      INT NOT NULL DEFAULT 0,
    processing  INT NOT NULL DEFAULT 0,
    sent        INT NOT NULL DEFAULT 0,
    failed      INT NOT NULL DEFAULT 0,
    cancelled   INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notifications (
    id              UUID PRIMARY KEY,
    batch_id        UUID REFERENCES batches(id) ON DELETE SET NULL,
    recipient       VARCHAR(255) NOT NULL,
    channel         VARCHAR(20)  NOT NULL,
    content         TEXT         NOT NULL,
    priority        VARCHAR(20)  NOT NULL DEFAULT 'normal',
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending',
    idempotency_key VARCHAR(255) UNIQUE,
    scheduled_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at         TIMESTAMPTZ,
    failed_at       TIMESTAMPTZ,
    retry_count     INT NOT NULL DEFAULT 0,
    provider_id     VARCHAR(255),
    error_message   TEXT
);

CREATE INDEX idx_notifications_batch_id    ON notifications(batch_id);
CREATE INDEX idx_notifications_status      ON notifications(status);
CREATE INDEX idx_notifications_channel     ON notifications(channel);
CREATE INDEX idx_notifications_created_at  ON notifications(created_at);
CREATE INDEX idx_notifications_scheduled   ON notifications(scheduled_at) WHERE scheduled_at IS NOT NULL;
