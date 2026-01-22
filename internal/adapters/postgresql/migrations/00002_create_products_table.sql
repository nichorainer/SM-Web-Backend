-- +goose Up
-- +goose StatementBegin
-- 00002_create_products_table.sql
CREATE TABLE IF NOT EXISTS products (
  id SERIAL PRIMARY KEY,
  product_id TEXT UNIQUE NOT NULL,    -- generated ID (e.g., P-0001)
  product_name TEXT NOT NULL,
  supplier_name TEXT NOT NULL,
  category TEXT NOT NULL,
  price_idr BIGINT NOT NULL DEFAULT 0, 
  stock INTEGER NOT NULL DEFAULT 0,
  created_by TEXT REFERENCES users(users_id) ON DELETE SET NULL,
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
