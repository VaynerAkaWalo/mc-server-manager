-- +goose Up
LOCK TABLE mc_servers IN SHARE ROW EXCLUSIVE MODE;

CREATE TABLE saved_server_names (
    name text PRIMARY KEY,
    save_id uuid NOT NULL REFERENCES saves(id) ON DELETE CASCADE
);

CREATE INDEX saved_server_names_save_id_idx ON saved_server_names (save_id);

INSERT INTO saved_server_names (name, save_id)
SELECT DISTINCT m.name, m.save_id FROM mc_servers m JOIN saves s ON s.id = m.save_id;

-- +goose StatementBegin
CREATE FUNCTION reserve_saved_server_name() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    reserved_save uuid;
BEGIN
    PERFORM pg_advisory_xact_lock(hashtextextended('server-name:' || NEW.name, 0));

    IF NEW.save_id IS NOT NULL THEN
        INSERT INTO saved_server_names (name, save_id) VALUES (NEW.name, NEW.save_id)
        ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
        WHERE saved_server_names.save_id = EXCLUDED.save_id
        RETURNING save_id INTO reserved_save;

        IF NOT FOUND THEN
            RAISE unique_violation USING MESSAGE = 'server name is reserved for another saved world', CONSTRAINT = 'saved_server_names_pkey';
        END IF;
    ELSIF EXISTS (SELECT 1 FROM saved_server_names WHERE name = NEW.name) THEN
        RAISE unique_violation USING MESSAGE = 'server name is reserved for a saved world', CONSTRAINT = 'saved_server_names_pkey';
    END IF;

    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER mc_servers_saved_name BEFORE INSERT ON mc_servers
FOR EACH ROW EXECUTE FUNCTION reserve_saved_server_name();

-- +goose Down
DROP TRIGGER mc_servers_saved_name ON mc_servers;
DROP FUNCTION reserve_saved_server_name();
DROP TABLE saved_server_names;
