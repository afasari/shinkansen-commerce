-- name: UpdateProduct :one
-- Update product fields (only update provided fields)
-- :id, :name, :description, :category_id, :price_units, :active
UPDATE catalog.products
SET
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    category_id = COALESCE(sqlc.narg('category_id'), category_id),
    price_units = COALESCE(sqlc.narg('price_units'), price_units),
    active = COALESCE(sqlc.narg('active'), active),
    updated_at = NOW()
WHERE id = sqlc.narg('id')
  AND deleted_at IS NULL
RETURNING id, name, description, category_id, price_units, price_currency, sku, active, stock_quantity, created_at, updated_at;
