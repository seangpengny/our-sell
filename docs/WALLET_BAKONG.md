# Wallet and Bakong top-ups

## What is implemented

Our Sell now has a USD wallet backed by a double-entry-style movement ledger:

- `wallets` stores the current `available_amount` and `held_amount` summary.
- `wallet_topups` stores each KHQR payment attempt and its provider reference.
- `wallet_ledger_entries` stores immutable balance movements and post-movement balances.
- Wallet Page orders atomically move USD from available funds to held funds.
- Expired wallet reservations return held funds to available funds and write an `order_release` entry.

The customer dashboard has **Wallet**. The admin dashboard has **Wallet payments** for searching and reviewing top-up status, customer, amount, and Bakong transaction ID.

## Runtime configuration

Copy the optional values from `.env.example` into the API environment:

```text
BAKONG_API_TOKEN=
BAKONG_API_BASE_URL=https://api-bakong.nbc.gov.kh
BAKONG_ACCOUNT_ID=
BAKONG_MERCHANT_NAME=Our Sell
BAKONG_MERCHANT_CITY=Phnom Penh
BAKONG_MERCHANT_ID=
BAKONG_ACQUIRING_BANK=
BAKONG_STORE_LABEL=Our Sell Wallet
BAKONG_TOPUP_TTL=15m
BAKONG_MIN_TOPUP_USD=1
BAKONG_MAX_TOPUP_USD=10000
BAKONG_WEBHOOK_SECRET=
BAKONG_HTTP_TIMEOUT=10s
```

Leave `BAKONG_API_TOKEN` and `BAKONG_ACCOUNT_ID` empty to keep the provider disabled. Production top-ups must use credentials and a merchant/account arrangement issued by a Bakong member bank or approved provider. Do not commit credentials.

## Payment flow

1. The authenticated customer enters a USD amount.
2. The API creates a unique `WT-...` reference, generates a dynamic KHQR, and stores its MD5 hash.
3. The dashboard renders the QR and polls the authenticated top-up endpoint.
4. The API verifies the payment server-side by QR MD5, recipient account, currency, amount, and provider transaction hash.
5. One database transaction credits the wallet and writes the `topup_confirmed` ledger entry. Repeated polling is idempotent.
6. A wallet Page order locks the wallet, decreases available funds, increases held funds, reserves the listing, and writes `order_hold` in one SQL statement.

The frontend never changes the balance. The backend is the only authority for crediting and holding funds.

## Status meanings

- `pending_payment`: QR created and awaiting a verified payment.
- `confirmed`: payment verified and USD credited exactly once.
- `failed`: provider data was returned but did not match the expected recipient, amount, currency, or hash.
- `expired`: the QR passed its expiry time before verification.
- `reversed`: reserved for a future provider reversal/reconciliation workflow.

## API surface

Authenticated customer endpoints:

- `GET /api/v1/wallet`
- `GET /api/v1/wallet/ledger?page=1&page_size=20`
- `POST /api/v1/wallet/topups` with `{ "amount_usd": 10.00 }`
- `GET /api/v1/wallet/topups/:topupID`

Admin endpoint:

- `GET /api/v1/admin/wallet/topups?status=confirmed&search=...`
- `POST /api/v1/admin/wallet/topups/:topupID/recheck`

The admin recheck action repeats the same server-side Bakong validation; it never manually credits a balance.

Optional internal relay endpoint:

- `POST /api/v1/payments/bakong/webhook` with `X-Bakong-Webhook-Secret` and either `reference` or `qr_md5`.

The public Bakong integration used here is the documented server-side transaction check by MD5. See the [Bakong Open API document](https://bakong.nbc.gov.kh/download/KHQR/integration/Bakong%20Open%20API%20Document.pdf).

## Before production

- Obtain and test the real Bakong credentials in a sandbox/SIT environment.
- Confirm whether the merchant account settles USD and that the configured account ID is the payment recipient.
- Configure HTTPS, secret management, request logging without tokens, and monitoring for failed/reversed payments.
- Run reconciliation tests for duplicate callbacks, delayed status, wrong amount, wrong recipient, expired QR, and provider timeout.
- Add a scheduled reconciliation worker before relying on polling alone.
- Add the final order completion/cancellation controls so held funds become spent or released according to the verified handover outcome.
