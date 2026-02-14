-- name: CreateOrder :one
INSERT INTO orders.orders (
    order_number, user_id, status,
    subtotal_units, subtotal_currency,
    tax_units, tax_currency,
    discount_units, discount_currency,
    total_units, total_currency,
    points_applied, shipping_address, payment_method
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING id;

-- name: GetOrder :one
SELECT id, order_number, user_id, status,
       subtotal_units, subtotal_currency,
       tax_units, tax_currency,
       discount_units, discount_currency,
       total_units, total_currency,
       points_applied, shipping_address, payment_method,
       created_at, updated_at
FROM orders.orders
WHERE id = $1;
