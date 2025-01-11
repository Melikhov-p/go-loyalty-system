-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "order"(
    id SERIAL PRIMARY KEY,
    number VARCHAR(100) NOT NULL UNIQUE,
    status VARCHAR(100) NOT NULL DEFAULT 'NEW',
    accrual NUMERIC(10, 2) NULL,
    uploaded_at TIMESTAMP,
    user_id INTEGER,
    FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "order";
-- +goose StatementEnd
