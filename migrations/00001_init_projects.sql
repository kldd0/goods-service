-- +goose Up
-- +goose StatementBegin

CREATE TABLE projects
(
    id         bigint PRIMARY KEY NOT NULL,
    name       text               NOT NULL,
    created_at timestamp          NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX projects_id_idx on projects (id);

INSERT INTO projects (id, name, created_at) VALUES (1, 'Запись 1', NOW());

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE projects;

-- +goose StatementEnd
