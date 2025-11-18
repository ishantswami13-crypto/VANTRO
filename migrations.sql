CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT NOT NULL,
    email       TEXT UNIQUE NOT NULL,
    api_key     TEXT UNIQUE NOT NULL
);

INSERT INTO users (name, email, api_key)
VALUES ('Admin', 'admin@example.com', 'supersecretapikey')
ON CONFLICT (api_key) DO NOTHING;

CREATE TABLE IF NOT EXISTS shops (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id     UUID NOT NULL REFERENCES users(id),
    name         TEXT NOT NULL,
    address      TEXT,
    gst_number   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS products (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id              UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name                 TEXT NOT NULL,
    sku                  TEXT,
    stock                INT NOT NULL DEFAULT 0,
    cost_price           NUMERIC(12,2) NOT NULL DEFAULT 0,
    selling_price        NUMERIC(12,2) NOT NULL DEFAULT 0,
    low_stock_threshold  INT NOT NULL DEFAULT 5
);

CREATE TABLE IF NOT EXISTS invoices (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id        UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    customer_name  TEXT,
    customer_phone TEXT,
    total_amount   NUMERIC(12,2) NOT NULL,
    tax_amount     NUMERIC(12,2) NOT NULL DEFAULT 0,
    status         TEXT NOT NULL DEFAULT 'PAID',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS invoice_items (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id  UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    product_id  UUID NOT NULL REFERENCES products(id),
    quantity    INT NOT NULL,
    unit_price  NUMERIC(12,2) NOT NULL
);

CREATE TABLE IF NOT EXISTS expenses (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id     UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    category    TEXT NOT NULL,
    amount      NUMERIC(12,2) NOT NULL,
    note        TEXT,
    spent_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS pots (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id        UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    target_amount  NUMERIC(12,2) NOT NULL,
    current_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);