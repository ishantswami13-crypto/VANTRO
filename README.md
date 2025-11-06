# VANTRO — Payouts API (Sandbox v1)

Developer-first payouts for UPI & bank transfers. Simple API, instant sandbox.

## Prereqs
- Go 1.22+
- Postgres 14+
- `psql`

## Setup

```bash
cp .env.example .env
export $(cat .env | xargs)

make migrate
make run
# -> listening on :8080

Test (sandbox)
API=sk_test_123456

# Create payout (UPI)
curl -s -X POST http://localhost:8080/v1/payouts \
  -H "Authorization: Bearer $API" -H "Content-Type: application/json" \
  -d '{
    "amount": 1250.5, "currency":"INR", "method":"upi",
    "upi": {"vpa":"rahul@upi","name":"Rahul"},
    "reference_id":"ORDER-92117"
  }' | jq .

# Get status (replace ID from create response)
curl -s -H "Authorization: Bearer $API" http://localhost:8080/v1/payouts/po_<id> | jq .

# Ledger
curl -s -H "Authorization: Bearer $API" "http://localhost:8080/v1/payouts/ledger?limit=20" | jq .

# Replay last webhook payload (simulated)
curl -s -X POST -H "Authorization: Bearer $API" http://localhost:8080/v1/payouts/po_<id>/webhook/replay | jq .

Notes

Sandbox provider returns processing then asynchronously sets success (~92%) or failed.

Webhooks are simulated via /webhook/replay.

In production, implement a real provider (Razorpay Payouts, bank) behind the Provider interface.
```
