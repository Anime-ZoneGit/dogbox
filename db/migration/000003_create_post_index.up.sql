BEGIN;

CREATE UNIQUE INDEX idx_posts_fn ON posts (filename);

COMMIT;
