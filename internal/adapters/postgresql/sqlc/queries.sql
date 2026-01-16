-- name: ListProducts :many
SELECT * FROM products;

--n name: FindProductByID :one
SELECT * FROM products WHERE id = $1;

-- name: GetOrderByID :one
SELECT 
  o.id AS order_id,
  o.created_at,
  c.name AS customer_name,
  SUM(p.price_in_idr * oi.quantity) AS total_price
FROM orders o
JOIN order_items oi ON o.id = oi.order_id
JOIN products p ON oi.product_id = p.id
JOIN customers c ON o.customer_id = c.id
WHERE o.id = $1
GROUP BY o.id, o.created_at, c.name;

-- name: CreateProduct :one
INSERT INTO products (name, price_in_idr, quantity)
VALUES ($1, $2, $3)
RETURNING id, name, price_in_idr, quantity, created_at;

-- name: GetProductByID :one
SELECT id, name, price_in_idr, quantity, created_at
FROM products
WHERE id = $1;