-- +goose Up
ALTER TABLE mc_servers ADD COLUMN owner_subject text;

UPDATE mc_servers m SET owner_subject = s.owner_subject FROM saves s WHERE m.save_id = s.id;

CREATE INDEX mc_servers_owner_created_at_id_idx ON mc_servers (owner_subject, created_at DESC, id DESC) WHERE owner_subject IS NOT NULL;

-- +goose StatementBegin
CREATE FUNCTION assign_saved_server_owner() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.owner_subject IS NULL AND NEW.save_id IS NOT NULL THEN
        SELECT owner_subject INTO NEW.owner_subject FROM saves WHERE id = NEW.save_id;
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER mc_servers_saved_owner BEFORE INSERT ON mc_servers
FOR EACH ROW EXECUTE FUNCTION assign_saved_server_owner();

-- +goose Down
DROP TRIGGER mc_servers_saved_owner ON mc_servers;
DROP FUNCTION assign_saved_server_owner();
DROP INDEX mc_servers_owner_created_at_id_idx;
ALTER TABLE mc_servers DROP COLUMN owner_subject;
