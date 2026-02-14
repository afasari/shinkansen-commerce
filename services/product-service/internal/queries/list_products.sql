-- name: ListProducts :many
-- List products with optional filtering
-- :category_id, :active_only, :limit, :offset
SELECT id, name, description, category_id, price_units, price_currency, sku, active, stock_quantity, created_at, updated_at
FROM catalog.products
WHERE deleted_at IS NULL
  AND (sqlc.narg('category_id')::uuid IS NULL OR category_id = sqlc.narg('category_id'))
  AND (sqlc.narg('active_only')::boolean IS NULL OR active = sqlc.narg('active_only'))
ORDER BY created_at DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');
