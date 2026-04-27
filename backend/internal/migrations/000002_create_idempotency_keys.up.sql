CREATE TABLE IF NOT EXISTS idempotency_keys (
    key          VARCHAR(255) PRIMARY KEY,
    request_hash VARCHAR(64)  NOT NULL,
    response     JSONB        NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
