-- name: EnsureWallet :one
INSERT INTO wallets (id, user_id, currency)
VALUES ($1, $2, 'USD')
ON CONFLICT (user_id) DO UPDATE SET updated_at = wallets.updated_at
RETURNING id, user_id, currency,
          available_amount::double precision AS available_amount,
          held_amount::double precision AS held_amount,
          created_at, updated_at;

-- name: GetWallet :one
SELECT id, user_id, currency,
       available_amount::double precision AS available_amount,
       held_amount::double precision AS held_amount,
       created_at, updated_at
FROM wallets
WHERE user_id = $1;

-- name: GetWalletForUpdate :one
SELECT id, user_id, currency,
       available_amount::double precision AS available_amount,
       held_amount::double precision AS held_amount,
       created_at, updated_at
FROM wallets
WHERE id = $1
FOR UPDATE;

-- name: GetWalletAvailableBalance :one
SELECT available_amount::double precision AS available_amount
FROM wallets
WHERE user_id = $1 AND currency = 'USD';

-- name: ListWalletLedger :many
SELECT id, wallet_id, user_id, entry_type, reference,
       available_delta_usd::double precision AS available_delta_usd,
       held_delta_usd::double precision AS held_delta_usd,
       available_balance_usd::double precision AS available_balance_usd,
       held_balance_usd::double precision AS held_balance_usd,
       description, created_at
FROM wallet_ledger_entries
WHERE user_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3;

-- name: CreateWalletTopup :one
INSERT INTO wallet_topups (
    id, wallet_id, user_id, reference, amount_usd, currency, qr_payload, qr_md5, expires_at
)
VALUES ($1, $2, $3, $4, $5::double precision, 'USD', $6::text, $7::varchar, $8::timestamptz)
RETURNING id, wallet_id, user_id, reference,
          amount_usd::double precision AS amount_usd, currency, qr_payload, qr_md5,
          provider_transaction_id, status, failure_reason, resolution_note,
          expires_at, paid_at, created_at, updated_at;

-- name: GetWalletTopup :one
SELECT id, wallet_id, user_id, reference,
       amount_usd::double precision AS amount_usd, currency, qr_payload, qr_md5,
       provider_transaction_id, status, failure_reason, resolution_note,
       expires_at, paid_at, created_at, updated_at
FROM wallet_topups
WHERE id = $1 AND user_id = $2;

-- name: GetWalletTopupByID :one
SELECT id, wallet_id, user_id, reference,
       amount_usd::double precision AS amount_usd, currency, qr_payload, qr_md5,
       provider_transaction_id, status, failure_reason, resolution_note,
       expires_at, paid_at, created_at, updated_at
FROM wallet_topups
WHERE id = $1;

-- name: GetWalletTopupForUpdate :one
SELECT id, wallet_id, user_id, reference,
       amount_usd::double precision AS amount_usd, currency, qr_payload, qr_md5,
       provider_transaction_id, status, failure_reason, resolution_note,
       expires_at, paid_at, created_at, updated_at
FROM wallet_topups
WHERE id = $1
FOR UPDATE;

-- name: GetWalletTopupByReference :one
SELECT id, wallet_id, user_id, reference,
       amount_usd::double precision AS amount_usd, currency, qr_payload, qr_md5,
       provider_transaction_id, status, failure_reason, resolution_note,
       expires_at, paid_at, created_at, updated_at
FROM wallet_topups
WHERE reference = $1;

-- name: GetWalletTopupByMD5 :one
SELECT id, wallet_id, user_id, reference,
       amount_usd::double precision AS amount_usd, currency, qr_payload, qr_md5,
       provider_transaction_id, status, failure_reason, resolution_note,
       expires_at, paid_at, created_at, updated_at
FROM wallet_topups
WHERE qr_md5 = $1;

-- name: CreditWallet :one
UPDATE wallets
SET available_amount = available_amount + $2::double precision,
    updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, currency,
          available_amount::double precision AS available_amount,
          held_amount::double precision AS held_amount,
          created_at, updated_at;

-- name: InsertWalletLedgerEntry :exec
INSERT INTO wallet_ledger_entries (
    id, wallet_id, user_id, entry_type, reference,
    available_delta_usd, held_delta_usd,
    available_balance_usd, held_balance_usd, description
)
VALUES ($1, $2, $3, $4::varchar, $5::varchar, $6::double precision, $7::double precision,
        $8::double precision, $9::double precision, $10::text);

-- name: ConfirmWalletTopup :one
UPDATE wallet_topups
SET status = 'confirmed',
    provider_transaction_id = COALESCE(NULLIF($2::varchar, ''), provider_transaction_id),
    paid_at = COALESCE(paid_at, NOW()),
    updated_at = NOW()
WHERE id = $1 AND status = 'pending_payment'
RETURNING id, wallet_id, user_id, reference,
          amount_usd::double precision AS amount_usd, currency, qr_payload, qr_md5,
          provider_transaction_id, status, failure_reason, resolution_note,
          expires_at, paid_at, created_at, updated_at;

-- name: MarkWalletTopupExpired :exec
UPDATE wallet_topups
SET status = 'expired', failure_reason = 'QR expired before payment was verified', updated_at = NOW()
WHERE id = $1 AND status = 'pending_payment';

-- name: MarkWalletTopupFailed :exec
UPDATE wallet_topups
SET status = 'failed', failure_reason = $2::text, updated_at = NOW()
WHERE id = $1 AND status = 'pending_payment';

-- name: ListWalletTopupsAdmin :many
SELECT t.id, t.wallet_id, t.user_id, u.name AS user_name, u.email AS user_email,
       t.reference, t.amount_usd::double precision AS amount_usd, t.currency,
       t.qr_md5, t.provider_transaction_id, t.status, t.failure_reason,
       t.resolution_note, t.expires_at, t.paid_at, t.created_at, t.updated_at
FROM wallet_topups t
JOIN users u ON u.id = t.user_id
WHERE ($1::text = '' OR t.reference ILIKE '%' || $1::text || '%' OR u.email ILIKE '%' || $1::text || '%')
  AND ($2::varchar = '' OR t.status = $2::varchar)
ORDER BY t.created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountWalletTopupsAdmin :one
SELECT COUNT(*)
FROM wallet_topups t
JOIN users u ON u.id = t.user_id
WHERE ($1::text = '' OR t.reference ILIKE '%' || $1::text || '%' OR u.email ILIKE '%' || $1::text || '%')
  AND ($2::varchar = '' OR t.status = $2::varchar);
