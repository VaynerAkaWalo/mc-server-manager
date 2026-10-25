-- +goose Up
ALTER TABLE mc_servers ADD COLUMN last_active_at timestamptz;

UPDATE mc_servers SET last_active_at = now() WHERE status IN ('running', 'degraded');

-- +goose Down
ALTER TABLE mc_servers DROP COLUMN last_active_at;
