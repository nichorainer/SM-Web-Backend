-- +goose Up
-- +goose StatementBegin
-- 00002_create_products_table.sql
CREATE TABLE IF NOT EXISTS products (
  id SERIAL PRIMARY KEY,
  product_id VARCHAR(50) UNIQUE NOT NULL,    -- generated ID (e.g., P-0001)
  product_name VARCHAR(255) NOT NULL,
  supplier_name VARCHAR(255),
  category VARCHAR(100),
  price_idr BIGINT NOT NULL DEFAULT 0,       -- store price in IDR as integer
  stock INTEGER NOT NULL DEFAULT 0,
  created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_products_product_id ON products(product_id);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products;
-- +goose StatementEnd
