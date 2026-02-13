-- name: GetProduct :one
-- Get a single product by ID
-- :product_id
SELECT
    id, name, description, category_id, price_units, price_currency,
    sku, active, stock_quantity, created_at, updated_at
FROM catalog.products
WHERE id = $1
  AND deleted_at IS NULL;
