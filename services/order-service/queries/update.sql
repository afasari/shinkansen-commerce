-- Name: update_order_status
-- Description: Update order status
-- Schema: orders

-- name: UpdateOrderStatus :exec
UPDATE orders.orders
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateOrderWithPoints :exec
UPDATE orders.orders
SET points_applied = $2, discount_units = $3, total_units = total_units - $3, updated_at = NOW()
WHERE id = $1;
