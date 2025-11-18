-- users (reference via supabase user_id or your own uuid)
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  email TEXT UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS expenses (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  amount_cents INTEGER NOT NULL CHECK (amount_cents >= 0),
  category TEXT NOT NULL,            -- food, travel, bills, shopping, misc
  mood TEXT,                         -- optional: calm, stressed...
  note TEXT,
  spent_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_expenses_user_time ON expenses(user_id, spent_at DESC);

CREATE TABLE IF NOT EXISTS saving_pots (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  target_cents INTEGER NOT NULL CHECK (target_cents >= 0),
  saved_cents INTEGER NOT NULL DEFAULT 0 CHECK (saved_cents >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION touch_updated_at() RETURNS trigger AS $$
BEGIN NEW.updated_at = now(); RETURN NEW; END; $$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_pots_updated_at ON saving_pots;
CREATE TRIGGER trg_pots_updated_at BEFORE UPDATE ON saving_pots
FOR EACH ROW EXECUTE PROCEDURE touch_updated_at();

CREATE TABLE IF NOT EXISTS coach_plans (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  week_start DATE NOT NULL,
  rules JSONB NOT NULL,              -- ["No food delivery on weekdays", ...]
  daily_nudge TEXT,                  -- “Breathe. Check pots before spending.”
  health_score INTEGER NOT NULL,     -- 0..100
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(user_id, week_start)
);
