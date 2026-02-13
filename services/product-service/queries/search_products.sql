-- name: SearchProducts :many
-- Full-text search for products using PostgreSQL GIN indexes
-- :query, :category_id, :min_price, :max_price, :in_stock_only, :limit, :offset
SELECT
    id, name, description, category_id, price_units, price_currency,
    sku, active, stock_quantity, created_at, updated_at,
    ts_rank(product_search_vector, plainto_tsquery($1::text)) AS rank
FROM catalog.products
WHERE deleted_at IS NULL
  AND ($2::uuid IS NULL OR category_id = $2)
  AND ($3::bigint IS NULL OR price_units >= $3)
  AND ($4::bigint IS NULL OR price_units <= $4)
  AND ($5::boolean IS NULL OR (in_stock_only = false OR stock_quantity > 0))
  AND product_search_vector @@ plainto_tsquery($1::text)
ORDER BY rank ASC, created_at DESC
LIMIT $6 OFFSET $7;
