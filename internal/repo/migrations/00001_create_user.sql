-- +goose Up
-- +goose StatementBegin

BEGIN;
CREATE TABLE IF NOT EXISTS "user"
(
    id          bigserial PRIMARY KEY,
    login       TEXT UNIQUE NOT NULL,
    password    TEXT NOT NULL,
    salt        TEXT NOT NULL,
    balance     double precision NOT NULL DEFAULT 0,
    withdrawn   double precision NOT NULL DEFAULT 0,
    created_at  timestamp NOT NULL DEFAULT NOW(),
    updated_at  timestamp NOT NULL DEFAULT NOW()
);
COMMIT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS "user";

-- +goose StatementEnd
