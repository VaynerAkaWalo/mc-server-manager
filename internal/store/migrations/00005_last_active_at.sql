-- +goose Up
ALTER TABLE mc_servers ADD COLUMN last_active_at timestamptz;

-- +goose Down
ALTER TABLE mc_servers DROP COLUMN last_active_at;
