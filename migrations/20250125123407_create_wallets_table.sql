-- +goose Up
-- +goose StatementBegin
CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    balance BIGINT NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE wallets;
-- +goose StatementEnd
