-- +goose Up
-- +goose StatementBegin
ALTER TABLE products ADD COLUMN low_stock_alert BOOLEAN DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products DROP COLUMN low_stock_alert;
-- +goose StatementEnd
