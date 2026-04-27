CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS expenses (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    amount      BIGINT      NOT NULL CHECK (amount > 0),
    category    VARCHAR(64) NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    date        DATE        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
