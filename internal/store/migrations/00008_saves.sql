-- +goose Up
CREATE TABLE saves (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_subject text NOT NULL,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX saves_owner_subject_created_at_id_idx ON saves (owner_subject, created_at, id);

ALTER TABLE mc_servers ADD COLUMN save_id uuid;
ALTER TABLE mc_servers ADD COLUMN save_locked boolean NOT NULL DEFAULT false;

CREATE UNIQUE INDEX mc_servers_save_lock_idx ON mc_servers (save_id) WHERE save_locked;
CREATE INDEX mc_servers_save_id_created_at_idx ON mc_servers (save_id, created_at DESC) WHERE save_id IS NOT NULL;

-- +goose Down
DROP INDEX mc_servers_save_id_created_at_idx;
DROP INDEX mc_servers_save_lock_idx;
ALTER TABLE mc_servers DROP COLUMN save_locked;
ALTER TABLE mc_servers DROP COLUMN save_id;
DROP INDEX saves_owner_subject_created_at_id_idx;
DROP TABLE saves;
