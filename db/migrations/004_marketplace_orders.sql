-- +goose Up

CREATE TABLE marketplace_orders (
    id UUID PRIMARY KEY,
    reference VARCHAR(32) NOT NULL UNIQUE,
    listing_id UUID NOT NULL REFERENCES facebook_page_listings(id) ON DELETE RESTRICT,
    buyer_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    buyer_name VARCHAR(100) NOT NULL,
    buyer_email TEXT NOT NULL,
    buyer_telegram VARCHAR(100) NOT NULL DEFAULT '',
    amount_cents BIGINT NOT NULL CHECK (amount_cents >= 0),
    currency VARCHAR(3) NOT NULL,
    payment_method VARCHAR(30) NOT NULL DEFAULT '',
    status VARCHAR(30) NOT NULL DEFAULT 'pending_payment',
    reservation_expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    CONSTRAINT marketplace_orders_status_check CHECK (
        status IN ('pending_payment', 'payment_failed', 'paid', 'transfer_pending', 'completed', 'cancelled', 'expired')
    )
);

CREATE INDEX marketplace_orders_listing_idx
    ON marketplace_orders (listing_id, status, reservation_expires_at);
CREATE INDEX marketplace_orders_status_idx
    ON marketplace_orders (status, created_at DESC);
CREATE INDEX marketplace_orders_buyer_email_idx
    ON marketplace_orders (buyer_email, created_at DESC);

-- +goose Down

DROP TABLE IF EXISTS marketplace_orders;
