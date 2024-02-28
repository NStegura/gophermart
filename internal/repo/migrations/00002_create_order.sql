-- +goose Up
-- +goose StatementBegin

BEGIN;
CREATE TYPE status_type AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE IF NOT EXISTS "order"
(
    id          bigserial PRIMARY KEY,
    status      status_type NOT NULL DEFAULT 'NEW',
    user_id     bigint NOT NULL,
    accrual     double precision NULL,
    created_at  timestamp NOT NULL DEFAULT NOW(),
    updated_at  timestamp NOT NULL DEFAULT NOW(),
    CONSTRAINT FK_order_user FOREIGN KEY(user_id) REFERENCES "user"(id)
                                                    ON DELETE RESTRICT
                                                    ON UPDATE CASCADE
);
CREATE INDEX idx_created_at ON "order"(created_at);
COMMIT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_created_at;
DROP TABLE IF EXISTS "order";
DROP TYPE IF EXISTS status_type;

-- +goose StatementEnd
