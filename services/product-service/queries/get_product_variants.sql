-- name: GetProductVariants :many
-- Get all variants for a specific product
-- :product_id
SELECT id, product_id, name, attributes, price_units, price_currency, sku, stock_quantity, created_at, updated_at
FROM catalog.product_variants
WHERE product_id = sqlc.narg('product_id')
ORDER BY created_at DESC;
