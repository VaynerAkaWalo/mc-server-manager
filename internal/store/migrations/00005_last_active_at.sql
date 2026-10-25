-- +goose Up
ALTER TABLE mc_servers ADD COLUMN last_active_at timestamptz;

UPDATE mc_servers SET last_active_at = now() WHERE status IN ('running', 'degraded');

DROP INDEX mc_servers_active_name_idx;
CREATE UNIQUE INDEX mc_servers_active_name_idx ON mc_servers (name) WHERE status IN ('new', 'running', 'degraded', 'expiring');

-- +goose Down
UPDATE mc_servers SET status = 'running' WHERE status = 'expiring';

DROP INDEX mc_servers_active_name_idx;
CREATE UNIQUE INDEX mc_servers_active_name_idx ON mc_servers (name) WHERE status IN ('new', 'running', 'degraded');

ALTER TABLE mc_servers DROP COLUMN last_active_at;
