-- +goose Up
-- +goose StatementBegin
ALTER TABLE orders
ADD COLUMN id_from_product INT REFERENCES products(id), -- FK ke PK products
ADD COLUMN product_id TEXT,                             -- kode produk (FPS001, dst.)
ADD COLUMN price_idr INTEGER DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE orders
DROP COLUMN IF EXISTS id_from_product,
DROP COLUMN IF EXISTS product_id,
DROP COLUMN IF EXISTS price_idr;
-- +goose StatementEnd