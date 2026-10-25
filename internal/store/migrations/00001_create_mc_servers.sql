-- +goose Up
CREATE TABLE mc_servers (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text        NOT NULL,
    hostname   text        NOT NULL,
    definition jsonb       NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL
);

CREATE INDEX mc_servers_name_idx ON mc_servers (name);

-- +goose Down
DROP TABLE mc_servers;
