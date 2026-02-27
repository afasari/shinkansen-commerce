-- name: SearchProductsFuzzy :many
-- Fuzzy search for products combining exact full-text match with trigram similarity
-- Supports typo correction and fuzzy matching
-- :search_query, :category_filter, :min_price, :max_price, :stock_only, :fuzzy_threshold
SELECT * FROM catalog.search_products_fuzzy(
    sqlc.narg('search_query'),
    sqlc.narg('category_filter'),
    sqlc.narg('min_price'),
    sqlc.narg('max_price'),
    sqlc.narg('stock_only'),
    sqlc.narg('fuzzy_threshold')
);
