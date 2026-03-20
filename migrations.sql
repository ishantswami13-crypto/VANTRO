-- Enable UUID generation (Postgres)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =========================
-- USERS
-- =========================
CREATE TABLE IF NOT EXISTS users (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email        TEXT UNIQUE NOT NULL,
  phone        TEXT UNIQUE,
  password_hash TEXT NOT NULL,

  full_name    TEXT,
  is_active    BOOLEAN NOT NULL DEFAULT TRUE,

  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- =========================
-- BUSINESSES
-- =========================
CREATE TABLE IF NOT EXISTS businesses (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  name         TEXT NOT NULL,
  industry     TEXT,
  currency     TEXT NOT NULL DEFAULT 'INR',

  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_businesses_owner_user_id
  ON businesses(owner_user_id);

-- =========================
-- ACCOUNTS (Cash/Bank/UPI/Wallet etc.)
-- =========================
CREATE TABLE IF NOT EXISTS accounts (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  business_id  UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,

  name         TEXT NOT NULL,
  type         TEXT NOT NULL, -- 'cash' | 'bank' | 'upi' | 'wallet' | 'other'
  currency     TEXT NOT NULL DEFAULT 'INR',
  opening_balance NUMERIC(14,2) NOT NULL DEFAULT 0,

  is_archived  BOOLEAN NOT NULL DEFAULT FALSE,

  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT accounts_type_check CHECK (type IN ('cash','bank','upi','wallet','other'))
);

CREATE INDEX IF NOT EXISTS idx_accounts_business_id
  ON accounts(business_id);

-- =========================
-- CATEGORIES (Expense/Income buckets)
-- =========================
CREATE TABLE IF NOT EXISTS categories (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  business_id  UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,

  name         TEXT NOT NULL,
  kind         TEXT NOT NULL, -- 'income' | 'expense'
  is_active    BOOLEAN NOT NULL DEFAULT TRUE,

  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT categories_kind_check CHECK (kind IN ('income','expense')),
  CONSTRAINT categories_unique_per_business UNIQUE (business_id, kind, name)
);

CREATE INDEX IF NOT EXISTS idx_categories_business_id
  ON categories(business_id);

-- =========================
-- TRANSACTIONS (the heart)
-- =========================
CREATE TABLE IF NOT EXISTS transactions (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  business_id  UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
  account_id   UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,

  category_id  UUID REFERENCES categories(id) ON DELETE SET NULL,

  direction    TEXT NOT NULL, -- 'income' | 'expense'
  amount       NUMERIC(14,2) NOT NULL,
  currency     TEXT NOT NULL DEFAULT 'INR',

  occurred_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

  description  TEXT,
  reference    TEXT, -- optional external reference id (UPI txn id etc.)

  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT transactions_direction_check CHECK (direction IN ('income','expense')),
  CONSTRAINT transactions_amount_positive CHECK (amount > 0)
);

CREATE INDEX IF NOT EXISTS idx_transactions_business_occurred_at
  ON transactions(business_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_transactions_account_occurred_at
  ON transactions(account_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_transactions_category_id
  ON transactions(category_id);

-- =========================
-- UPDATED_AT auto-touch (optional, but clean)
-- =========================
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_users_updated_at'
  ) THEN
    CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_businesses_updated_at'
  ) THEN
    CREATE TRIGGER trg_businesses_updated_at
    BEFORE UPDATE ON businesses
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_accounts_updated_at'
  ) THEN
    CREATE TRIGGER trg_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_transactions_updated_at'
  ) THEN
    CREATE TRIGGER trg_transactions_updated_at
    BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
END IF;
END $$;

-- =========================
-- DEFAULT CATEGORY SEED (optional)
-- =========================
DO $$
DECLARE
  biz RECORD;
  income_categories TEXT[] := ARRAY['Sales', 'Other Income'];
  expense_categories TEXT[] := ARRAY['Rent', 'Supplies', 'Marketing', 'Travel', 'Food', 'Utilities', 'Salary', 'Other Expense'];
BEGIN
  FOR biz IN SELECT id FROM businesses LOOP
    FOREACH name IN ARRAY income_categories LOOP
      INSERT INTO categories (business_id, name, kind)
      VALUES (biz.id, name, 'income')
      ON CONFLICT (business_id, kind, name) DO NOTHING;
    END LOOP;

    FOREACH name IN ARRAY expense_categories LOOP
      INSERT INTO categories (business_id, name, kind)
      VALUES (biz.id, name, 'expense')
      ON CONFLICT (business_id, kind, name) DO NOTHING;
    END LOOP;
  END LOOP;
END $$;

-- =========================
-- Subscriptions (Razorpay)
-- =========================
CREATE TABLE IF NOT EXISTS subscriptions (
  id BIGSERIAL PRIMARY KEY,
  client_id TEXT NOT NULL UNIQUE,
  status TEXT NOT NULL DEFAULT 'inactive',
  expires_at TIMESTAMPTZ,
  reference_id TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_client_status ON subscriptions (client_id, status);
