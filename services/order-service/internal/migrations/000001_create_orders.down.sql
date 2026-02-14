-- Name: create_orders
-- Description: Drop orders table and schema

DROP TABLE IF EXISTS orders.order_items CASCADE;
DROP TABLE IF EXISTS orders.orders CASCADE;
DROP SCHEMA IF EXISTS orders CASCADE;
