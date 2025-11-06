CREATE TABLE IF NOT EXISTS payouts (
  id TEXT PRIMARY KEY,
  reference_id TEXT,
  amount_cents BIGINT NOT NULL,
  currency TEXT NOT NULL,
  method TEXT NOT NULL, -- 'upi' | 'bank'
  dest_vpa TEXT,
  dest_name TEXT,
  dest_account TEXT,
  dest_ifsc TEXT,
  status TEXT NOT NULL, -- 'processing'|'success'|'failed'
  utr TEXT,
  error TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS payout_events (
  id TEXT PRIMARY KEY,
  payout_id TEXT NOT NULL,
  event TEXT NOT NULL, -- payout.processing|payout.success|payout.failed
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_payouts_ref ON payouts(reference_id);
CREATE INDEX IF NOT EXISTS idx_payouts_status ON payouts(status);
