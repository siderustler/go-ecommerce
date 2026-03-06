-- 01_schema.sql
-- Base schema for go-ecommerce.

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    main_image TEXT NOT NULL,
    price INT NOT NULL,
    category TEXT CHECK (category IN ('MACHINES', 'GARDENING', 'ELECTRO', 'PARTS', 'ELECTROMACHINES')),
    price_before INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS product_images (
    id UUID REFERENCES products(id),
    image_url TEXT NOT NULL,
    PRIMARY KEY (id, image_url)
);

CREATE TABLE IF NOT EXISTS product_details (
    id SERIAL PRIMARY KEY,
    product_id UUID REFERENCES products(id),
    product_info TEXT NOT NULL,
    technical_parameters TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS billings (
    id UUID PRIMARY KEY,
    nip_code TEXT,
    company TEXT,
    city TEXT,
    address TEXT,
    postal_code TEXT,
    local TEXT
);

CREATE TABLE IF NOT EXISTS shippings (
    id UUID PRIMARY KEY,
    city TEXT,
    address TEXT,
    postal_code TEXT,
    local TEXT
);

CREATE TABLE IF NOT EXISTS customers (
    customer_id TEXT PRIMARY KEY,
    name TEXT,
    email TEXT,
    phone TEXT,
    billing UUID REFERENCES billings(id),
    shipping UUID REFERENCES shippings(id)
);

CREATE TABLE IF NOT EXISTS baskets (
    id UUID PRIMARY KEY,
    customer_id TEXT NOT NULL,
    last_modified_at TIMESTAMP NOT NULL,
    status TEXT CHECK (status IN ('ACTIVE', 'INACTIVE'))
);

CREATE TABLE IF NOT EXISTS checkouts (
    id UUID PRIMARY KEY,
    customer_id TEXT NOT NULL REFERENCES customers(customer_id),
    basket_id UUID NOT NULL REFERENCES baskets(id),
    created_at TIMESTAMP NOT NULL,
    status TEXT CHECK (status IN ('INVALIDATED', 'PENDING', 'FINALIZED'))
);

CREATE TABLE IF NOT EXISTS basket_products (
    id UUID NOT NULL REFERENCES baskets(id),
    product_id UUID NOT NULL REFERENCES products(id),
    count INT NOT NULL,
    PRIMARY KEY (id, product_id)
);

CREATE TABLE IF NOT EXISTS stock (
    product_id UUID PRIMARY KEY NOT NULL REFERENCES products(id),
    available_amount INT DEFAULT 0,
    reserved_amount INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS stock_reservations (
    checkout_id UUID REFERENCES checkouts(id),
    product_id UUID REFERENCES products(id),
    amount INT DEFAULT 0,
    reserved_at TIMESTAMP NOT NULL,
    PRIMARY KEY (checkout_id, product_id)
);

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY,
    checkout_id UUID REFERENCES checkouts(id),
    created_at TIMESTAMP NOT NULL,
    status TEXT CHECK (status IN ('FINALIZED', 'PAID', 'SHIPPING'))
);

CREATE TABLE IF NOT EXISTS order_products (
    id SERIAL PRIMARY KEY,
    order_id UUID REFERENCES orders(id),
    name TEXT NOT NULL,
    price REAL NOT NULL,
    count INT NOT NULL
);
