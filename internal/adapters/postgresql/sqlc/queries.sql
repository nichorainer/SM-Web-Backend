-- name: ListProducts :many
SELECT * FROM products;

--n name: FindProductByID :one
SELECT * FROM products WHERE id = $1;