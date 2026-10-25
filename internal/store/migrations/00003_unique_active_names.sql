-- +goose Up
UPDATE mc_servers SET status = 'expired' WHERE status IN ('new', 'running') AND expires_at <= now();

UPDATE mc_servers older SET status = 'expired'
WHERE older.status IN ('new', 'running')
  AND EXISTS (
    SELECT 1 FROM mc_servers newer
    WHERE newer.name = older.name
      AND newer.status IN ('new', 'running')
      AND (newer.created_at, newer.id) > (older.created_at, older.id)
  );

CREATE UNIQUE INDEX mc_servers_active_name_idx ON mc_servers (name) WHERE status IN ('new', 'running');

-- +goose Down
DROP INDEX mc_servers_active_name_idx;
