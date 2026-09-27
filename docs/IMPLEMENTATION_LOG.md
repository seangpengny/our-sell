# Implementation log

## 2026-09-25 — Wallet-first Page checkout

- Made the signed-in Page checkout use the USD wallet as its only currently supported order payment method.
- Added a server-side available-balance check in the checkout UI before order creation, with the backend remaining the final atomic authority.
- Added a clear insufficient-balance state and a link to the Wallet top-up screen.
- Added a safe `return_to` flow so a confirmed Bakong top-up sends the buyer back to the original checkout.
- Rejected non-wallet Page order payment methods in the API until their provider integrations are implemented.
- Updated checkout copy and marketplace documentation to describe wallet holds instead of pending external payment.

## 2026-09-25 — USD wallet and Bakong top-ups

- Added migration `007_wallet_and_bakong.sql` for USD wallets, top-up records, and immutable ledger entries.
- Added server-side KHQR generation using a unique top-up reference and QR MD5.
- Added Bakong transaction verification by MD5, recipient/amount/currency validation, and idempotent wallet crediting.
- Added authenticated Wallet APIs and customer Wallet UI with a QR code, polling, balance summary, and ledger history.
- Added wallet payment to Page checkout; available funds move to held funds atomically with the order reservation.
- Added automatic held-fund release when an unpaid wallet reservation expires.
- Added admin Wallet payments search and status review.
- Added `WALLET_BAKONG.md` with configuration, operating rules, status meanings, and production checklist.

## 2026-09-25 — USD marketplace amounts

- Added migration `006_usd_amounts.sql` to rename listing `price_cents` to `price_usd` and order `amount_cents` to `amount_usd`.
- Converted existing cent values to exact PostgreSQL `NUMERIC(12,2)` USD amounts.
- Enforced `USD` as the only marketplace currency and validated prices to two decimal places.
- Updated SQLC queries, API models, OpenAPI, admin listing management, storefront pricing, and checkout display to use USD amounts.

## 2026-09-25 — Authenticated marketplace orders

- Added migration `005_marketplace_order_buyer_identity.sql`.
- Marketplace orders now use the authenticated `buyer_user_id` and one optional `buyer_note`.
- Removed duplicated `buyer_name`, `buyer_email`, and `buyer_telegram` columns from new order data.
- Protected order creation with authentication; the server derives the buyer ID from the bearer token.
- Updated checkout to require sign-in and return the buyer to checkout after login.
- Preserved details from the existing legacy test order in `buyer_note` before removing the old columns.
- Kept the legacy order’s nullable buyer ID for history; the database check rejects missing buyer IDs on new orders.

## 2026-09-25 — Orders and listing reservations

Completed the next marketplace phase:

- Added `marketplace_orders` in migration `004_marketplace_orders.sql`.
- Stored the buyer identity, optional buyer note, price snapshot, currency, order reference, status, and reservation expiry.
- Added an atomic reservation query that changes a listing from `published` to `reserved` while creating the order.
- Added a 15-minute `pending_payment` reservation window.
- Added automatic expiry handling: expired reservations become `expired` and the listing returns to `published`.
- Added `POST /api/v1/marketplace/pages/:listingID/orders`.
- Added server-side validation for the buyer note and terms acceptance.
- Added an authenticated checkout call before the payment step.
- Updated checkout confirmation to show the real order reference and reservation expiry.
- Kept payment controls as preview-only; no money is collected.
- Regenerated SQLC and applied database migration version 4 locally.

## 2026-09-24 — Page inventory and listings

- Added persistent Facebook Page inventory tables and Page-to-connection relationships.
- Added marketplace listing records with draft, review, published, reserved, sold, and transfer-oriented statuses.
- Added admin Page Management with search, account filtering, pagination, synchronization, and listing editing.
- Added public APIs for published Page listings.
- Updated the frontend marketplace and detail pages to read published listings from PostgreSQL.

## 2026-09-24 — Integration fixes

- Added `PUT` and `PATCH` to the CORS method allow-list.
- Removed the UI-only `price` field from listing-save payloads; the API receives `price_usd` only.
- Added an explicit PostgreSQL status cast in the listing upsert query.
- Added administrator protection to Page inventory, synchronization, and listing-management endpoints.

## Verification

The following checks have passed during implementation:

```text
go test ./...
npm run typecheck
npm run lint
npm run build
make migrate-up
```

The local database is currently migrated through version 7.

## Not implemented yet

- Direct Bakong checkout for Page orders; Bakong is currently implemented for wallet top-ups.
- A scheduled reconciliation worker; top-up polling and the optional signed relay endpoint are implemented.
- Automatic listing reservation cleanup worker; expired reservations are currently released when a marketplace read or reservation request runs.
- Admin order dashboard and final order-status controls for converting held funds to spent or released on verified handover.
- Seller/buyer notifications.
- Meta Page ownership-transfer automation.
