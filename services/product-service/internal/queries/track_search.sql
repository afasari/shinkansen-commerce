-- name: TrackSearch :exec
-- Track search query to analytics table for business intelligence
-- Records search query, results count, and optional user_id
-- :search_query, :results_count, :user_id
SELECT catalog.track_search(
    sqlc.narg('search_query'),
    sqlc.narg('results_count'),
    sqlc.narg('user_id')
);
