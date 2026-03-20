-- products table
CREATE TABLE IF NOT EXISTS products (
  id            UUID PRIMARY KEY,
  name          TEXT        NOT NULL,
  sku           TEXT        UNIQUE,
  price_cents   INTEGER     NOT NULL CHECK (price_cents >= 0),
  currency      TEXT        NOT NULL DEFAULT 'INR',
  stock         INTEGER     NOT NULL DEFAULT 0,
  active        BOOLEAN     NOT NULL DEFAULT TRUE,
  metadata      JSONB       NOT NULL DEFAULT '{}'::jsonb,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- quick update trigger to keep updated_at fresh
CREATE OR REPLACE FUNCTION touch_updated_at() RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_products_updated_at ON products;
CREATE TRIGGER trg_products_updated_at
BEFORE UPDATE ON products
FOR EACH ROW EXECUTE PROCEDURE touch_updated_at();
