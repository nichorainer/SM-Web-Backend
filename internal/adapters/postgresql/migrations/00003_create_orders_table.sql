-- +goose Up
-- +goose StatementBegin
-- 00003_create_orders_table.sql
CREATE TABLE IF NOT EXISTS orders (
  id SERIAL PRIMARY KEY,
  order_number TEXT UNIQUE NOT NULL,
  customer_name TEXT NOT NULL,
  total_amount INTEGER DEFAULT 1,
  status TEXT NOT NULL,
  platform TEXT NOT NULL,
  destination TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- CREATE INDEX IF NOT EXISTS idx_orders_order_number ON orders(order_number);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
