BEGIN;

CREATE OR REPLACE FUNCTION pos_by_id(id bigint) RETURNS bigint AS $$
    SELECT COUNT(id) FROM public."posts" WHERE id <= $1;
$$ LANGUAGE SQL IMMUTABLE;

CREATE INDEX idx_posts_by_id_pos ON posts USING btree(pos_by_id(id));

COMMIT;
