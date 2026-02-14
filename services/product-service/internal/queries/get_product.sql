-- name: GetProduct :one
-- Get a single product by ID
-- :id
SELECT id, name, description, category_id, price_units, price_currency, sku, active, stock_quantity, created_at, updated_at
FROM catalog.products
WHERE id = sqlc.narg('id')
  AND deleted_at IS NULL;
