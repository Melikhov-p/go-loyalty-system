-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS watched_order(
    id SERIAL PRIMARY KEY ,
    order_number VARCHAR(100) NOT NULL UNIQUE ,
    FOREIGN KEY (order_number) REFERENCES "order"(number) ON DELETE CASCADE ,
    user_id INTEGER,
    FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE ,
    accrual_order_status VARCHAR(100) NOT NULL DEFAULT 'NEW',
    trackable BOOLEAN DEFAULT true
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS watched_order;
-- +goose StatementEnd
