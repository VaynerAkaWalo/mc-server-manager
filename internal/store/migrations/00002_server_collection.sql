-- +goose Up
ALTER TABLE mc_servers ADD COLUMN status text NOT NULL DEFAULT 'provisioning'
    CHECK (status IN ('provisioning', 'running', 'stopped', 'expired'));

CREATE INDEX mc_servers_created_at_id_idx ON mc_servers (created_at, id);
CREATE INDEX mc_servers_expires_at_id_idx ON mc_servers (expires_at, id);
CREATE INDEX mc_servers_status_created_at_id_idx ON mc_servers (status, created_at, id);
CREATE INDEX mc_servers_status_expires_at_id_idx ON mc_servers (status, expires_at, id);

-- +goose Down
DROP INDEX mc_servers_status_expires_at_id_idx;
DROP INDEX mc_servers_status_created_at_id_idx;
DROP INDEX mc_servers_expires_at_id_idx;
DROP INDEX mc_servers_created_at_id_idx;
ALTER TABLE mc_servers DROP COLUMN status;
