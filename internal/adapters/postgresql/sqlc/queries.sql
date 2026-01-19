-- name: ListProducts :many
SELECT id, name, price_in_idr, quantity, created_at
FROM products
ORDER BY created_at DESC;

-- name: FindProductByID :one
SELECT id, name, price_in_idr, quantity, created_at
FROM products
WHERE id = $1;

-- name: CreateProduct :one
INSERT INTO products (name, price_in_idr, quantity)
VALUES ($1, $2, $3)
RETURNING id, name, price_in_idr, quantity, created_at;

-- name: UpdateProductStock :exec
UPDATE products
SET quantity = $2
WHERE id = $1;

-- name: GetOrderByID :one
SELECT 
  o.id AS order_id,
  o.created_at,
  c.name AS customer_name,
  COALESCE(SUM(p.price_in_idr * oi.quantity), 0) AS total_price
FROM orders o
JOIN customers c ON o.customer_id = c.id
LEFT JOIN order_items oi ON o.id = oi.order_id
LEFT JOIN products p ON oi.product_id = p.id
WHERE o.id = $1
GROUP BY o.id, o.created_at, c.name;

-- name: CreateOrder :one
INSERT INTO orders (customer_id)
VALUES ($1)
RETURNING id, customer_id, created_at;

-- name: AddOrderItem :exec
INSERT INTO order_items (order_id, product_id, quantity)
VALUES ($1, $2, $3);

-- name: ListCustomers :many
SELECT id, name, email, created_at
FROM customers
ORDER BY created_at DESC;

-- name: FindCustomerByID :one
SELECT id, name, email, created_at
FROM customers
WHERE id = $1;

-- name: CreateCustomer :one
INSERT INTO customers (name, email)
VALUES ($1, $2)
RETURNING id, name, email, created_at;

-- name: ListLowStockProducts :many
SELECT id, name, quantity
FROM products
WHERE quantity < 5;