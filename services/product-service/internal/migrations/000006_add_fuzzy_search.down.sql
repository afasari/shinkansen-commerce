DROP FUNCTION IF EXISTS catalog.get_top_search_queries(integer);
DROP FUNCTION IF EXISTS catalog.track_search(text, integer, uuid);
DROP TABLE IF EXISTS catalog.search_analytics;
DROP FUNCTION IF EXISTS catalog.search_products_fuzzy(text, uuid, bigint, bigint, boolean, real);
DROP INDEX IF EXISTS catalog.idx_products_name_trgm;
