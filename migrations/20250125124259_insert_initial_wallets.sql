-- +goose Up
-- +goose StatementBegin
INSERT INTO wallets (balance) VALUES
   (0),
   (20000),
   (0),
   (40000),
   (50000);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM wallets;
-- +goose StatementEnd
