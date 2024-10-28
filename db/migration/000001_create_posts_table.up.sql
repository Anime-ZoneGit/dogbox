BEGIN;

CREATE TYPE post_status AS ENUM ('pending', 'ok', 'removed');

CREATE TABLE IF NOT EXISTS posts (
  id bigserial PRIMARY KEY,
  filename text CONSTRAINT filename_unique UNIQUE,
  deletion_key text,
  hash text CONSTRAINT hash_unique UNIQUE,
  status post_status DEFAULT 'pending',
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT has_data CHECK (
    status <> 'ok'
    OR (
      filename IS NOT NULL
      AND hash IS NOT NULL
    )
  )
);

COMMIT;
