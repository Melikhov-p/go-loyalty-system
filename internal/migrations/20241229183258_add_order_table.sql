-- +goose Up
-- +goose StatementBegin
CREATE TYPE status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS "order"(
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    number VARCHAR(100) NOT NULL UNIQUE,
    status status DEFAULT 'NEW',
    accrual NUMERIC(10, 2) NULL,
    uploaded_at TIMESTAMP NOT NULL,
    user_id INTEGER,
    FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS status;

DROP TABLE IF EXISTS "order";
-- +goose StatementEnd
