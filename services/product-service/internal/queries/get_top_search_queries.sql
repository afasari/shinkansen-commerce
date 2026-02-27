-- name: GetTopSearchQueries :many
-- Retrieve most popular search queries from the last N days
-- Returns top 100 queries with search count and unique user count
-- :days_ago
SELECT * FROM catalog.get_top_search_queries(
    sqlc.narg('days_ago')
);
