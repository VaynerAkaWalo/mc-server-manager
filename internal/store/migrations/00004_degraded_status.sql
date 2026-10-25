-- +goose Up
ALTER TABLE mc_servers DROP CONSTRAINT mc_servers_status_check;
ALTER TABLE mc_servers ADD CONSTRAINT mc_servers_status_check
    CHECK (status IN ('new', 'running', 'degraded', 'stopped', 'expired'));

DROP INDEX mc_servers_active_name_idx;
CREATE UNIQUE INDEX mc_servers_active_name_idx ON mc_servers (name) WHERE status IN ('new', 'running', 'degraded');

-- +goose Down
UPDATE mc_servers SET status = 'running' WHERE status = 'degraded';

DROP INDEX mc_servers_active_name_idx;
CREATE UNIQUE INDEX mc_servers_active_name_idx ON mc_servers (name) WHERE status IN ('new', 'running');

ALTER TABLE mc_servers DROP CONSTRAINT mc_servers_status_check;
ALTER TABLE mc_servers ADD CONSTRAINT mc_servers_status_check
    CHECK (status IN ('new', 'running', 'stopped', 'expired'));
