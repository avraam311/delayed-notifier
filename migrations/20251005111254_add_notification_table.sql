-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS notification (
    id SERIAL PRIMARY KEY,
    message TEXT NOT NULL,
    date_time TIMESTAMP NOT NULL,
    status VARCHAR(30),
    mail TEXT NOT NULL,
    tg_id INTEGER
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS notification;

-- +goose StatementEnd
