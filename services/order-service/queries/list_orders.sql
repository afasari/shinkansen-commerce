-- name: ListUserOrders :many
SELECT id, order_number, user_id, status, 
       subtotal_units, subtotal_currency,
       tax_units, tax_currency,
       discount_units, discount_currency,
       total_units, total_currency,
       points_applied, shipping_address, payment_method,
       created_at, updated_at
FROM orders.orders
WHERE user_id = $1
AND ($2::int4 IS NULL OR status = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;
