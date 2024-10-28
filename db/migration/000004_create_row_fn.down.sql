BEGIN;

DROP FUNCTION pos_by_id (id bigint);

DROP INDEX idx_posts_by_id_pos;

COMMIT;
