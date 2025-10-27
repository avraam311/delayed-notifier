-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS delete (
    id SERIAL PRIMARY KEY,
    delete_id INT NOT NULL UNIQUE
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS delete;

-- +goose StatementEnd