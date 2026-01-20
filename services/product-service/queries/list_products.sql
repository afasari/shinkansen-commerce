-- name: ListProducts
-- List products with optional filtering
-- :category_id, :active_only, :limit, :offset
SELECT
    id, name, description, category_id, price_units, price_currency,
    sku, active, stock_quantity, created_at, updated_at
FROM catalog.products
WHERE deleted_at IS NULL
  AND ($1::uuid IS NULL OR category_id = $1)
  AND ($2::boolean IS NULL OR active = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;
