-- +goose Up
-- +goose StatementBegin

CREATE TABLE logs (
    id          UInt64,
    project_id  UInt64,
    name        String,
    description Nullable(String),
    priority    Nullable(Int32),
    removed     Boolean,
    event_time  DateTime
) ENGINE = NATS
  SETTINGS nats_url = 'nats:4222',
           nats_subjects = 'clickhouse_logs',
           nats_format = 'JSONEachRow',
           date_time_input_format = 'best_effort';

CREATE INDEX logs_idx ON logs(id, project_id, name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS logs;

-- +goose StatementEnd
