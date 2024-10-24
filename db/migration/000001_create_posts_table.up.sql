CREATE TABLE IF NOT EXISTS posts (
    id bigserial PRIMARY KEY,
    filename text,
    deletion_key text,
    hash text,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_filename UNIQUE (filename),
    CONSTRAINT unique_hash UNIQUE(hash)
);
