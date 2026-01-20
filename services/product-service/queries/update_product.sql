-- name: UpdateProduct
-- Update product fields (only update provided fields)
-- :id, :name, :description, :category_id, :price_units, :active
UPDATE catalog.products
SET
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    category_id = COALESCE($4, category_id),
    price_units = COALESCE($5, price_units),
    active = COALESCE($6, active),
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;
