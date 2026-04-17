-- Seed data for Shinkansen Commerce
-- Run with: make db-seed
-- or: psql "$DATABASE_URL" -f scripts/seed-data.sql
--
-- Prerequisites: all migrations must be applied (make db-migrate)

-- ============================================================
-- 1. Delivery Zone (Kanto region)
-- ============================================================
INSERT INTO delivery.delivery_zones (id, name, postal_codes, prefectures, delivery_days)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'Kanto Region',
    ARRAY['100-0000','150-0000','160-0000','170-0000','180-0000','200-0000','210-0000','220-0000','230-0000','240-0000','250-0000','260-0000','270-0000','280-0000','290-0000','300-0000','310-0000','350-0000'],
    ARRAY['東京都','神奈川県','埼玉県','千葉県','茨城県','栃木県','群馬県'],
    1
) ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 2. Delivery Slots (today + 2 days, 4 slots per day)
-- ============================================================
INSERT INTO delivery.delivery_slots (id, delivery_zone_id, start_time, end_time, capacity, reserved, date)
VALUES
    ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001',
     CURRENT_DATE + INTERVAL '8 hours',  CURRENT_DATE + INTERVAL '10 hours', 10, 0, CURRENT_DATE),
    ('b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000001',
     CURRENT_DATE + INTERVAL '10 hours', CURRENT_DATE + INTERVAL '12 hours', 10, 0, CURRENT_DATE),
    ('b0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000001',
     CURRENT_DATE + INTERVAL '14 hours', CURRENT_DATE + INTERVAL '16 hours', 10, 0, CURRENT_DATE),
    ('b0000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000001',
     CURRENT_DATE + INTERVAL '16 hours', CURRENT_DATE + INTERVAL '18 hours', 10, 0, CURRENT_DATE),
    ('b0000000-0000-0000-0000-000000000005', 'a0000000-0000-0000-0000-000000000001',
     CURRENT_DATE + INTERVAL '1 day 8 hours',  CURRENT_DATE + INTERVAL '1 day 10 hours',  10, 0, CURRENT_DATE + INTERVAL '1 day'),
    ('b0000000-0000-0000-0000-000000000006', 'a0000000-0000-0000-0000-000000000001',
     CURRENT_DATE + INTERVAL '1 day 10 hours', CURRENT_DATE + INTERVAL '1 day 12 hours', 10, 0, CURRENT_DATE + INTERVAL '1 day'),
    ('b0000000-0000-0000-0000-000000000007', 'a0000000-0000-0000-0000-000000000001',
     CURRENT_DATE + INTERVAL '1 day 14 hours', CURRENT_DATE + INTERVAL '1 day 16 hours', 10, 0, CURRENT_DATE + INTERVAL '1 day'),
    ('b0000000-0000-0000-0000-000000000008', 'a0000000-0000-0000-0000-000000000001',
     CURRENT_DATE + INTERVAL '1 day 16 hours', CURRENT_DATE + INTERVAL '1 day 18 hours', 10, 0, CURRENT_DATE + INTERVAL '1 day'),
    ('b0000000-0000-0000-0000-000000000009', 'a0000000-0000-0000-0000-000000000001',
     CURRENT_DATE + INTERVAL '2 days 8 hours',  CURRENT_DATE + INTERVAL '2 days 10 hours',  10, 0, CURRENT_DATE + INTERVAL '2 days'),
    ('b0000000-0000-0000-0000-000000000010', 'a0000000-0000-0000-0000-000000000001',
     CURRENT_DATE + INTERVAL '2 days 10 hours', CURRENT_DATE + INTERVAL '2 days 12 hours', 10, 0, CURRENT_DATE + INTERVAL '2 days'),
    ('b0000000-0000-0000-0000-000000000011', 'a0000000-0000-0000-0000-000000000001',
     CURRENT_DATE + INTERVAL '2 days 14 hours', CURRENT_DATE + INTERVAL '2 days 16 hours', 10, 0, CURRENT_DATE + INTERVAL '2 days'),
    ('b0000000-0000-0000-0000-000000000012', 'a0000000-0000-0000-0000-000000000001',
     CURRENT_DATE + INTERVAL '2 days 16 hours', CURRENT_DATE + INTERVAL '2 days 18 hours', 10, 0, CURRENT_DATE + INTERVAL '2 days')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 3. Product Categories
-- ============================================================
INSERT INTO catalog.categories (id, name, level)
VALUES
    ('c0000000-0000-0000-0000-000000000001', 'Electronics',  0),
    ('c0000000-0000-0000-0000-000000000002', 'Clothing',     0),
    ('c0000000-0000-0000-0000-000000000003', 'Food & Drink', 0)
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 4. Products
-- ============================================================
INSERT INTO catalog.products (id, name, description, category_id, sku, price_units, price_currency, active, stock_quantity)
VALUES
    ('d0000000-0000-0000-0000-000000000001',
     'Shinkansen Model Train N700S',
     'Highly detailed 1:200 scale N700S Shinkansen bullet train model with LED headlights and motorized mechanism.',
     'c0000000-0000-0000-0000-000000000001',
     'SKS-MDL-N700S-001', 5800, 'JPY', true, 50),

    ('d0000000-0000-0000-0000-000000000002',
     'Tokaido Shinkansen T-Shirt',
     'Premium cotton t-shirt featuring a stylish Tokaido Shinkansen route map design. Available in multiple sizes.',
     'c0000000-0000-0000-0000-000000000002',
     'SKS-TSH-TOKAI-001', 3200, 'JPY', true, 100),

    ('d0000000-0000-0000-0000-000000000003',
     'Ekiben Bento Box Set',
     'Authentic Japanese ekiben (station bento) experience with a curated selection of regional specialties. Ships frozen.',
     'c0000000-0000-0000-0000-000000000003',
     'SKS-FD-EBN-001', 2500, 'JPY', true, 30)
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 5. Inventory Stock Items (warehouse ID is a well-known constant)
-- ============================================================
-- This warehouse_id matches DEFAULT_WAREHOUSE_ID in the frontend
INSERT INTO inventory.stock_items (id, product_id, variant_id, warehouse_id, quantity, reserved_quantity)
VALUES
    ('e0000000-0000-0000-0000-000000000001',
     'd0000000-0000-0000-0000-000000000001',
     NULL,
     'f0000000-0000-0000-0000-000000000001',
     50, 0),
    ('e0000000-0000-0000-0000-000000000002',
     'd0000000-0000-0000-0000-000000000002',
     NULL,
     'f0000000-0000-0000-0000-000000000001',
     100, 0),
    ('e0000000-0000-0000-0000-000000000003',
     'd0000000-0000-0000-0000-000000000003',
     NULL,
     'f0000000-0000-0000-0000-000000000001',
     30, 0)
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Summary
-- ============================================================
-- Delivery Zone:  Kanto Region (a0000000-...)
-- Delivery Slots: 12 slots across 3 days (b0000000-...)
-- Categories:     Electronics, Clothing, Food & Drink (c0000000-...)
-- Products:       3 products (d0000000-...) at ¥2500–¥5800
-- Stock Items:    3 items in warehouse f0000000-... with 30–100 units each
