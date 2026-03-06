-- 02_seed.sql
-- Test-only seed data. Apply only in testing/development environments.

INSERT INTO products (id, name, main_image, price, category, price_before) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Electric Trimmer', '/public/products/essa/1.webp', 12999, 'MACHINES', 9999),
    ('00000000-0000-0000-0000-000000000002', 'Garden Shears', '/public/products/essa/2.webp', 4999, 'GARDENING', 0),
    ('00000000-0000-0000-0000-000000000003', 'Chainsaw Chain', '/public/products/essa/3.webp', 2599, 'PARTS', 0)
ON CONFLICT (id) DO NOTHING;

INSERT INTO product_images (id, image_url) VALUES
    ('00000000-0000-0000-0000-000000000001', '/public/products/essa/1.webp'),
    ('00000000-0000-0000-0000-000000000001', '/public/products/essa/2.webp'),
    ('00000000-0000-0000-0000-000000000001', '/public/products/essa/3.webp'),
    ('00000000-0000-0000-0000-000000000002', '/public/products/essa/1.webp'),
    ('00000000-0000-0000-0000-000000000003', '/public/products/essa/1.webp')
ON CONFLICT (id, image_url) DO NOTHING;

INSERT INTO product_details (product_id, product_info, technical_parameters) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Lightweight electric trimmer suitable for home gardens.', 'Power: 800W; Voltage: 230V'),
    ('00000000-0000-0000-0000-000000000002', 'Manual shears for precise cuts.', 'Blade length: 20cm; Weight: 420g'),
    ('00000000-0000-0000-0000-000000000003', 'Durable replacement chain for chainsaws.', 'Pitch: 3/8"; Gauge: 1.3mm')
ON CONFLICT DO NOTHING;

INSERT INTO stock (product_id, available_amount, reserved_amount) VALUES
    ('00000000-0000-0000-0000-000000000001', 100, 0),
    ('00000000-0000-0000-0000-000000000002', 100, 0),
    ('00000000-0000-0000-0000-000000000003', 100, 0)
ON CONFLICT (product_id) DO UPDATE
SET available_amount = EXCLUDED.available_amount,
    reserved_amount = EXCLUDED.reserved_amount;

INSERT INTO customers (customer_id, name, email, phone) VALUES
    ('test-user-1', 'Test User', 'test@example.com', '+48111111111')
ON CONFLICT (customer_id) DO NOTHING;
