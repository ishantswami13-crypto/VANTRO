# VANTRO MoneyOS — v0.1.0 (2025-11-08)

## What’s new

✅ Live Go API on Render (Fiber + Neon Postgres)

✅ Expenses: create + list

✅ Saving Pots: create, list, add funds

✅ Money Coach: weekly plan (rules + score)

✅ Apple-style UI polish (cards, pill buttons)

✅ JWT auth via Supabase (magic link)

✅ Web proxy (Next.js) to keep keys server-side

## Fixes & Performance

- CORS, gzip, rate limit middleware
- Faster cold-start by lazy DB init

## Setup/Deploy

Env: DATABASE_URL, PORT=10000, SUPABASE_JWKS_URL

Mobile run:

```
flutter run -d chrome --dart-define API_BASE=... --dart-define SUPABASE_URL=... --dart-define SUPABASE_ANON=...
```

## Known issues

- Expense listing paging not implemented
- Users upsert on first request (no profiles UI yet)
- Coach score is heuristic (LLM coach planned)
