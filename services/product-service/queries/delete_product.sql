-- name: DeleteProduct :exec
-- Soft delete a product by setting deleted_at timestamp
-- :id
UPDATE catalog.products
SET deleted_at = NOW(), active = false
WHERE id = $1;
