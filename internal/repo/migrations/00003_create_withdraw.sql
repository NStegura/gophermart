-- +goose Up
-- +goose StatementBegin

BEGIN;
CREATE TABLE IF NOT EXISTS "withdraw"
(
    id          bigserial PRIMARY KEY,
    order_id    bigint NOT NULL,
    user_id     bigint NOT NULL,
    sum         double precision NOT NULL,
    created_at  timestamp NOT NULL DEFAULT NOW(),
    CONSTRAINT FK_balance_user FOREIGN KEY(user_id) REFERENCES "user"(id)
                                                    ON DELETE RESTRICT
                                                    ON UPDATE CASCADE
);
COMMIT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS "withdraw";

-- +goose StatementEnd
