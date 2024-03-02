-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS goods (
    id          bigint PRIMARY KEY NOT NULL,
    project_id  bigint             NOT NULL,
    name        text               NOT NULL,
    description text,
    priority    int,
    removed     boolean                     DEFAULT FALSE,
    created_at  timestamp          NOT NULL DEFAULT NOW(),

    FOREIGN KEY (project_id) REFERENCES projects (id)
);

CREATE UNIQUE INDEX goods_id_name_idx on goods (id, project_id, name);


-- a trigger to automatically set the default value for the priority column
CREATE OR REPLACE FUNCTION priority()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.priority IS NULL THEN
        NEW.priority := COALESCE((SELECT MAX(priority) + 1 FROM goods), 1);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_default_priority
    BEFORE INSERT ON goods
    FOR EACH ROW
    EXECUTE FUNCTION priority();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS set_default_priority;

DROP FUNCTION IF EXISTS priority;

DROP TABLE IF EXISTS goods;

-- +goose StatementEnd
