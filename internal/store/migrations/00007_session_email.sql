-- +goose Up
ALTER TABLE auth_sessions ADD COLUMN email text NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE auth_sessions DROP COLUMN email;
