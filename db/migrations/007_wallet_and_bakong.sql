-- +goose Up

CREATE TABLE wallets (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    available_amount NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (available_amount >= 0),
    held_amount NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (held_amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT wallets_currency_usd_check CHECK (currency = 'USD')
);

CREATE TABLE wallet_topups (
    id UUID PRIMARY KEY,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    reference VARCHAR(32) NOT NULL UNIQUE,
    amount_usd NUMERIC(12,2) NOT NULL CHECK (amount_usd > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    qr_payload TEXT NOT NULL,
    qr_md5 VARCHAR(32) NOT NULL UNIQUE,
    provider_transaction_id VARCHAR(128) UNIQUE,
    status VARCHAR(24) NOT NULL DEFAULT 'pending_payment',
    failure_reason TEXT NOT NULL DEFAULT '',
    resolution_note TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NOT NULL,
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT wallet_topups_amount_currency_check CHECK (currency = 'USD'),
    CONSTRAINT wallet_topups_status_check CHECK (
        status IN ('pending_payment', 'confirmed', 'failed', 'expired', 'reversed')
    )
);

CREATE INDEX wallet_topups_user_idx ON wallet_topups (user_id, created_at DESC);
CREATE INDEX wallet_topups_status_idx ON wallet_topups (status, expires_at);

CREATE TABLE wallet_ledger_entries (
    id UUID PRIMARY KEY,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    entry_type VARCHAR(32) NOT NULL,
    reference VARCHAR(64) NOT NULL,
    available_delta_usd NUMERIC(12,2) NOT NULL,
    held_delta_usd NUMERIC(12,2) NOT NULL,
    available_balance_usd NUMERIC(12,2) NOT NULL CHECK (available_balance_usd >= 0),
    held_balance_usd NUMERIC(12,2) NOT NULL CHECK (held_balance_usd >= 0),
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT wallet_ledger_entry_type_check CHECK (
        entry_type IN ('topup_confirmed', 'topup_reversed', 'order_hold', 'order_release', 'order_spent', 'admin_adjustment')
    )
);

CREATE INDEX wallet_ledger_wallet_idx ON wallet_ledger_entries (wallet_id, created_at DESC);
CREATE INDEX wallet_ledger_user_idx ON wallet_ledger_entries (user_id, created_at DESC);

-- +goose Down

DROP TABLE IF EXISTS wallet_ledger_entries;
DROP TABLE IF EXISTS wallet_topups;
DROP TABLE IF EXISTS wallets;
