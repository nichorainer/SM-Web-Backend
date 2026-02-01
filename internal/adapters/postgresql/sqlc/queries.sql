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
RETURNING id, full_name, username, email;

-- Products

-- name: CreateProduct :one
INSERT INTO products (product_id, product_name, supplier_name, category, price_idr, stock, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, product_id, product_name, supplier_name, category, price_idr, stock, created_by, created_at, updated_at;

-- name: GetProductByProductID :one
SELECT id, product_id, product_name, supplier_name, category, price_idr, stock, created_by, created_at, updated_at
FROM products
WHERE product_id = $1
LIMIT 1;

-- name: ListProducts :many
SELECT id, product_id, product_name, supplier_name, category, price_idr, stock, created_by, created_at, updated_at
FROM products
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateProduct :one
UPDATE products
SET product_name = COALESCE(NULLIF($2, ''), product_name),
    supplier_name = COALESCE(NULLIF($3, ''), supplier_name),
    category = COALESCE(NULLIF($4, ''), category),
    price_idr = COALESCE($5, price_idr),
    stock = COALESCE($6, stock),
    updated_at = now()
WHERE product_id = $1
RETURNING id, product_id, product_name, supplier_name, category, price_idr, stock, created_by, created_at, updated_at;

-- Orders

-- name: CreateOrder :one
INSERT INTO orders (order_number, customer_name, total_amount, status, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, order_number, customer_name, total_amount, status, created_by, created_at, updated_at;

-- name: ListOrders :many
SELECT 
  id, 
  order_number, 
  customer_name, 
  total_amount, 
  status, 
  created_by, 
  created_at, 
  updated_at
FROM orders
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: GetOrderByID :one
SELECT id, order_number, customer_name, total_amount, status, created_by, created_at, updated_at
FROM orders
WHERE order_number = $1
LIMIT 1;

-- Utility queries

-- name: NextProductSequence :one
-- This is a helper to get a next sequence number for product id generation if you prefer DB-side sequence.
SELECT nextval('products_id_seq') as seq;