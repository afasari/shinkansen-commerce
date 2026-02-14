-- name: CreateProduct :one
-- Create a new product in the catalog
-- :name, :description, :category_id, :price_units, :price_currency, :sku, :stock_quantity
INSERT INTO catalog.products (name, description, category_id, price_units, price_currency, sku, stock_quantity, created_at, updated_at)
VALUES (sqlc.narg('name'), sqlc.narg('description'), sqlc.narg('category_id'), sqlc.narg('price_units'), sqlc.narg('price_currency'), sqlc.narg('sku'), sqlc.narg('stock_quantity'), NOW(), NOW())
RETURNING id;
