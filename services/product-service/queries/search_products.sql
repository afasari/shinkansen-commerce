-- name: SearchProducts :many
-- Full-text search for products using PostgreSQL GIN indexes
-- :query, :category_id, :min_price, :max_price, :in_stock_only, :limit, :offset
SELECT id, name, description, category_id, price_units, price_currency, sku, active, stock_quantity, created_at, updated_at,
       ts_rank(product_search_vector, plainto_tsquery(sqlc.narg('query'))) AS rank
FROM catalog.products
WHERE deleted_at IS NULL
  AND (sqlc.narg('category_id')::uuid IS NULL OR category_id = sqlc.narg('category_id'))
  AND (sqlc.narg('min_price')::bigint IS NULL OR price_units >= sqlc.narg('min_price'))
  AND (sqlc.narg('max_price')::bigint IS NULL OR price_units <= sqlc.narg('max_price'))
  AND (sqlc.narg('in_stock_only')::boolean IS NULL OR (in_stock_only = false OR stock_quantity > 0))
  AND product_search_vector @@ plainto_tsquery(sqlc.narg('query'))
ORDER BY rank ASC, created_at DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');
