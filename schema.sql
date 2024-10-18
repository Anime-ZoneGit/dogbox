CREATE TABLE posts (
    id  BIGSERIAL PRIMARY KEY,
    filename text NOT NULL,
    identifier text NOT NULL,
    uploaddate timestamptz NOT NULL,
    deletetoken text NOT NULL
);
