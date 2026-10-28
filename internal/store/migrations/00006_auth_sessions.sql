-- +goose Up
CREATE TABLE login_attempts (
    state_hash text PRIMARY KEY,
    verifier text NOT NULL,
    nonce text NOT NULL,
    return_to text NOT NULL,
    expires_at timestamptz NOT NULL
);

CREATE TABLE auth_sessions (
    id uuid NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    token_hash text PRIMARY KEY,
    subject text NOT NULL,
    expires_at timestamptz NOT NULL
);

CREATE INDEX login_attempts_expiry ON login_attempts (expires_at);
CREATE INDEX auth_sessions_expiry ON auth_sessions (expires_at);

-- +goose Down
DROP TABLE auth_sessions;
DROP TABLE login_attempts;
