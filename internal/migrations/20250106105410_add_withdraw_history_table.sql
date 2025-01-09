-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS withdraw_history (
    id SERIAL PRIMARY KEY ,
    order_number BIGINT NOT NULL ,
    sum numeric(10, 2),
    processed_at TIMESTAMP,
    user_id INTEGER,
    FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS withdraw_history;
-- +goose StatementEnd
