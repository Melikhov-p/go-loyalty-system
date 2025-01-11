-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS balance(
    id SERIAL PRIMARY KEY,
    current NUMERIC(10,2) DEFAULT 0,
    withdrawn NUMERIC(10,2) DEFAULT 0,
    user_id INTEGER,
    FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS balance;
-- +goose StatementEnd
