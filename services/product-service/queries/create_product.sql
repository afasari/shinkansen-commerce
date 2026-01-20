-- name: CreateProduct
-- Create a new product in the catalog
-- :product_id, :name, :description, :category_id, :price_units, :price_currency, :sku, :stock_quantity
INSERT INTO catalog.products (id, name, description, category_id, price_units, price_currency, sku, stock_quantity, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
RETURNING id;
