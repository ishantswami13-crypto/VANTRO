# VANTRO.Pvt

Optional alternate stack for VANTRO.

This folder contains:

- A Go API with runtime entrypoint `cmd/api/main.go`
- A Flutter client with entrypoint `lib/main.dart`

It is not required for the main web deployment, which uses `../vantro-ui` and `../vantro-backend`.

## API Environment

Use `.env.example` as the template.

- `PORT`
- `ENV`
- `DATABASE_URL`
- `API_KEY`
- `JWT_SECRET`
- `PROVIDER`
- `RAZORPAY_KEY_ID`
- `RAZORPAY_KEY_SECRET`
- `RAZORPAY_WEBHOOK_SECRET`

## Commands

- `go run ./cmd/api`
- `go test ./...`
- `go build ./...`
- `flutter pub get`
- `flutter analyze`

## Keep This Stack Only If

- You still support the Flutter/mobile client
- You still need this alternate backend implementation
