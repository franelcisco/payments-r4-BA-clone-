# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this service is

A Go (Gin) HTTP service that brokers the **R4Bank** payment API for Venezuelan mobile-payment flows (BCV exchange rate, OTP generation, immediate debit, direct debit by account/phone, change/vuelto). It also receives webhooks from R4Bank (`R4consulta`, `R4notifica`) and persists confirmed payments to Postgres via GORM.

The service is **multi-tenant by store**: it serves two independent commerces — `bone` and `appa` — each with its own R4Bank credentials, HMAC secret, route prefix, and database tables. The two share the same handler/service code; tenant selection is done at wiring time in `cmd/main.go` (separate `RestClient`, `R4Service`, and `WebhookAuthMiddleware` instances per store).

## Build, run, deploy

```bash
go build -o bin/server ./cmd        # local build
go run ./cmd                         # run (reads .env via godotenv autoload)
go mod tidy                          # update deps
```

There are **no tests** in the repo yet. `go test ./...` is a no-op.

`Dockerfile` builds a distroless image from `./cmd`. `deploy.sh` is the prod deploy script — it pulls `main`, rebuilds, runs `setcap` for low ports, and restarts the `goapp` systemd unit. Logs go to journald (`journalctl -u goapp`).

Required env vars (validated at startup in `internal/config/config.go`, service refuses to boot if any are missing):
- `PORT` (defaults to 8080)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `SSL_MODE`
- `R4_BONE_ENTRY_POINT`, `R4_BONE_COMMERCE_TOKEN`, `BONE_SECRET`
- `R4_APPA_ENTRY_POINT`, `R4_APPA_COMMERCE_TOKEN`, `APPA_SECRET`

## Architecture

Standard layered Go layout: `cmd → routers → handlers → services → pkg`. All dependencies are wired explicitly in `cmd/main.go`; there is no DI container.

```
cmd/main.go                       wires two parallel tenant stacks (bone + appa)
internal/routers/                 Gin route groups, per tenant
internal/handlers/                JSON binding + status code mapping only
internal/services/                business logic (R4Service, WebhookService)
internal/models/                  API request/response DTOs
pkg/r4bank/                       outbound HTTP client to R4Bank + DTOs
pkg/middleware/webhook_auth.go    inbound HMAC auth
pkg/db/                           gorm setup + table models + schema.sql
pkg/logs/                         zap logger
pkg/ipfy/                         logs public IP at startup
```

### Route layout (tenant prefix matters)

- `bone`: `/r4/*` (outbound API surface) + `/R4consulta`, `/R4notifica` (inbound webhooks)
- `appa`: `/r4/appa/*` (outbound API surface) + `/appa/R4consulta`, `/appa/R4notifica`
- `GET /healthz` — public liveness

The `appa` outbound route set is intentionally a subset — it omits `direct-debit-account` and `direct-debit-phone`. Keep these asymmetries in mind when adding endpoints.

### HMAC authentication (both directions, same algorithm)

`pkg/r4bank/client.go` defines `GenerateAuthToken(key, message)` = `hex(HMAC_SHA256(key, message))`. This is used in three places:

1. **Inbound webhook auth** (`pkg/middleware/webhook_auth.go`): the `Authorization` header must equal `HMAC(secret, commerceToken)`. Each tenant has its own secret + token pair, so the middleware is instantiated twice in `main.go`.
2. **Outbound R4Bank requests** (`RestClient.Do`): every call signs a per-endpoint `hmacInput` string with the commerce token and sends it as `Authorization`; `Commerce` header carries the token itself.
3. **Startup self-check** in `r4bank.NewClient`: if the constant-time HMAC compare fails it logs an error and returns **nil** — the resulting nil client will panic on first use. If you change token wiring, verify the client isn't nil before deploying.

The per-endpoint `hmacInput` is a positional concatenation of payload fields and **the order is part of the contract with R4Bank**. See each method in `internal/services/r4.go` (e.g. `ChangePaid` uses `Phone+Amount+Bank+DNI`, `GenerateOTP` uses `Bank+Amount+Phone+DNI`, `ValidateImmediateDebit` uses `Bank+DNI+Phone+Amount+OTP`). Reordering these will break auth silently.

### Webhook async semantics

- `R4consulta` (preview): handler returns `200 {status:true}` immediately and writes to DB in a **fire-and-forget goroutine**. Errors are only logged. This is deliberate — R4Bank expects a fast ACK.
- `R4notifica` (confirmation): synchronous DB write; **deduplicates on `reference`** (`existReference` in `WebhookService`) so retries are idempotent. Banks that send a 3-digit `BancoEmisor` are normalized to 4 digits with a leading zero before storage.

### Per-tenant DB tables

Schema lives in `pkg/db/schema.sql` (not auto-migrated — apply manually):
- `r4_mobile_payments` / `r4_mobile_payments_previews` → `bone`
- `r4_appa_mobile_payments` / `r4_appa_mobile_payments_previews` → `appa`

`WebhookService` switches on a `storeName` string (`"bone"` / `"appa"`) to pick the right table/GORM model. Adding a third tenant means: new env vars, new routes file, new GORM models + schema, and new `case` branches in `webhookService.{RegisterR4MobilePayment,RegisterR4MobilePaymentPreview,existReference}`.

All payment timestamps are written in **`America/Caracas`** (loaded once in `main.go` and passed into `WebhookService`).

### Long-running R4Bank operations

Two patterns coexist for waiting on async R4Bank operations:

- `ValidateImmediateDebit`: up to **7 polls** of `GetOperationByID` with `3s` sleeps; breaks when code ≠ `"AC00"` (in-progress sentinel). Looks up the final code in `_DebitInmediateSpecialResponse` for a human-readable message.
- `DirectDebitAccount`: `getDirectDebitResponse` ticks every **2s for up to 120s** (`context.WithTimeout`) until code ≠ `"AC00"`.

If you add another endpoint that returns an async operation ID, reuse `getDirectDebitResponse` rather than rolling a third polling loop.

## Conventions worth knowing

- Module name uses snake_case (`bone_appetit_r4_service`), not the usual kebab-case — imports look unusual but are correct.
- Error responses from handlers use `{"error": "..."}` for the `/r4/*` surface and `{"status": false}` / `{"abono": false}` for webhooks and auth failures — match the existing shape per surface.
- Logging is `zap` everywhere **except** handlers and a few service spots that still use `fmt.Printf`. Prefer `zap` for new code.
- Amounts arriving from R4Bank are strings (`Monto`); the service parses them with `strconv.ParseFloat`. Amounts sent to R4Bank are formatted with `fmt.Sprintf("%.2f", ...)` — keep this format, it's part of the HMAC input.
