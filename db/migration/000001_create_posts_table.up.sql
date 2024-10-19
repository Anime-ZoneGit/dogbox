CREATE TABLE IF NOT EXISTS posts (
    id bigserial PRIMARY KEY,
    filename text NOT NULL,
    deletion_key text NOT NULL,
    md5 text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
