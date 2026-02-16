-- internal/adapters/postgresql/sqlc/queries.sql

-- Users

-- name: CreateUser :one
INSERT INTO users (user_id, username, email, full_name, password_hash, role)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, username, email, full_name, role, created_at, updated_at;

-- name: ListUsers :many
SELECT user_id, username, email, full_name, role, created_at
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetUserByUsernameOrEmail :one
SELECT id, user_id, username, email, full_name, password_hash, role, created_at, updated_at
FROM users
WHERE username = $1 OR email = $2
LIMIT 1;

-- name: UserByID :one
SELECT id, user_id, username, email, full_name, password_hash, role, created_at, updated_at
FROM users
WHERE id = $1
LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET username = COALESCE(NULLIF($2, ''), username),
    email = COALESCE(NULLIF($3, ''), email),
    full_name = COALESCE(NULLIF($4, ''), full_name),
    password_hash = COALESCE(NULLIF($5, ''), password_hash),
    updated_at = now()
WHERE id = $1
RETURNING id, full_name, username, email, password_hash;

-- name: UpdateUserPermissions :exec
UPDATE users
SET permissions = $2, updated_at = now()
WHERE id = $1;

-- Products

-- name: CreateProduct :one
INSERT INTO products (product_id, product_name, supplier_name, category, price_idr, stock)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, product_id, product_name, supplier_name, category, price_idr, stock, created_at, updated_at;

-- name: GetProductByID :one
SELECT
  id,
  product_id,
  product_name,
  supplier_name,
  category,
  price_idr,
  stock,
  created_at,
  updated_at
FROM products
WHERE id = $1
LIMIT 1;

-- name: ListProducts :many
SELECT
  id,
  product_id,
  product_name,
  supplier_name,
  category,
  price_idr,
  stock,
  created_at,
  updated_at
FROM products
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateProductStock :one
UPDATE products
SET stock = $2,
    updated_at = now()
WHERE id = $1
RETURNING id, product_id, product_name, supplier_name, category, price_idr, stock, created_at, updated_at;

-- name: UpdateProductStockByDelta :one
UPDATE products
SET stock = stock + $2,
    updated_at = now()
WHERE id = $1
  AND (stock + $2) >= 0
RETURNING id, product_id, product_name, supplier_name, category, price_idr, stock, created_at, updated_at;

-- name: UpdateProduct :one
UPDATE products
SET product_id    = $2,
    product_name  = $3,
    supplier_name = $4,
    category      = $5,
    price_idr     = $6,
    stock         = $7,
    updated_at    = now()
WHERE id = $1
RETURNING id, product_id, product_name, supplier_name, category, price_idr, stock, created_at, updated_at;

-- Orders

-- name: CreateOrder :one
INSERT INTO orders (
    order_number,
    id_from_product,
    product_id,
    customer_name,
    platform,
    destination,
    total_amount,
    status,
    price_idr,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id = $1;

-- name: GetLastOrderNumber :one
SELECT order_number
FROM orders
ORDER BY id DESC
LIMIT 1;

-- name: ListOrdersWithProduct :many
SELECT
  o.id,
  o.order_number,
  o.id_from_product,       -- FK ke products.id
  o.product_id,            -- kode produk (text)
  o.customer_name,
  o.total_amount,
  o.status,
  o.platform,
  o.destination,
  o.price_idr,             -- harga produk
  o.created_at,
  p.product_name
FROM orders o
JOIN products p ON o.id_from_product = p.id
ORDER BY o.id DESC
LIMIT $1 OFFSET $2;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $2
WHERE id = $1
RETURNING *;

-- name: DeleteOrder :exec
DELETE FROM orders
WHERE id = $1;

-- Utility queries

-- name: NextProductSequence :one
-- This is a helper to get a next sequence number for product id generation if you prefer DB-side sequence.
SELECT nextval('products_id_seq') as seq;