-- +goose Up
-- +goose StatementBegin

BEGIN;
CREATE TYPE status_type AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE IF NOT EXISTS "order"
(
    id          bigserial PRIMARY KEY,
    status      status_type NOT NULL,
    user_id     bigint NOT NULL,
    created_at  timestamp NOT NULL DEFAULT NOW(),
    CONSTRAINT FK_order_user FOREIGN KEY(user_id) REFERENCES "user"(id)
                                                                ON DELETE RESTRICT
                                                                ON UPDATE CASCADE
);
COMMIT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS "order";
DROP TYPE IF EXISTS status_type;

-- +goose StatementEnd
