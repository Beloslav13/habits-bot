-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id            SERIAL PRIMARY KEY,
    telegram_id   BIGINT UNIQUE NOT NULL,
    first_name    VARCHAR(128)  NOT NULL,
    last_name     VARCHAR(128),
    username      VARCHAR(64),
    language_code VARCHAR(8),
    is_premium    BOOLEAN       NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_telegram_id ON users (telegram_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
