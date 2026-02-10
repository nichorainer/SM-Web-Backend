-- +goose Up
-- +goose StatementBegin
ALTER TABLE orders
ADD COLUMN price_idr INTEGER DEFAULT 0,
ADD COLUMN product_id INT REFERENCES products(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE orders
DROP COLUMN IF EXISTS price_idr,
DROP COLUMN IF EXISTS product_id;
-- +goose StatementEnd