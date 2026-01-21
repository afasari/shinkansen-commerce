-- Name: order_items
-- Description: Manage order items
-- Schema: orders

-- name: AddOrderItem :exec
INSERT INTO orders.order_items (
    order_id, product_id, variant_id, product_name, quantity,
    unit_price_units, unit_price_currency, total_price_units, total_price_currency
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (order_id, product_id) DO NOTHING;

-- name: GetOrderItems :many
SELECT id, order_id, product_id, variant_id, product_name, quantity,
       unit_price_units, unit_price_currency, total_price_units, total_price_currency, created_at
FROM orders.order_items
WHERE order_id = $1
ORDER BY created_at;

-- name: GetOrderItem :one
SELECT id, order_id, product_id, variant_id, product_name, quantity,
       unit_price_units, unit_price_currency, total_price_units, total_price_currency, created_at
FROM orders.order_items
WHERE id = $1;
