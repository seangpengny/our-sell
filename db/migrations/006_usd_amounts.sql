-- +goose Up

ALTER TABLE facebook_page_listings
    DROP CONSTRAINT IF EXISTS facebook_page_listings_price_cents_check;

ALTER TABLE facebook_page_listings
    RENAME COLUMN price_cents TO price_usd;

ALTER TABLE facebook_page_listings
    ALTER COLUMN price_usd TYPE NUMERIC(12,2)
    USING price_usd::numeric / 100;

UPDATE facebook_page_listings
SET currency = 'USD';

ALTER TABLE facebook_page_listings
    ADD CONSTRAINT facebook_page_listings_price_usd_check CHECK (price_usd >= 0),
    ADD CONSTRAINT facebook_page_listings_currency_usd_check CHECK (currency = 'USD');

ALTER TABLE marketplace_orders
    DROP CONSTRAINT IF EXISTS marketplace_orders_amount_cents_check;

ALTER TABLE marketplace_orders
    RENAME COLUMN amount_cents TO amount_usd;

ALTER TABLE marketplace_orders
    ALTER COLUMN amount_usd TYPE NUMERIC(12,2)
    USING amount_usd::numeric / 100;

-- Existing orders are already stored as USD. Do not rewrite legacy rows here:
-- migration 005 intentionally preserves one historical order without a buyer ID,
-- and PostgreSQL re-checks all row constraints when an UPDATE touches that row.

ALTER TABLE marketplace_orders
    ADD CONSTRAINT marketplace_orders_amount_usd_check CHECK (amount_usd >= 0),
    ADD CONSTRAINT marketplace_orders_currency_usd_check CHECK (currency = 'USD');

-- +goose Down

ALTER TABLE marketplace_orders
    DROP CONSTRAINT IF EXISTS marketplace_orders_amount_usd_check,
    DROP CONSTRAINT IF EXISTS marketplace_orders_currency_usd_check;

ALTER TABLE marketplace_orders
    ALTER COLUMN amount_usd TYPE BIGINT
    USING ROUND(amount_usd * 100)::bigint;

ALTER TABLE marketplace_orders
    RENAME COLUMN amount_usd TO amount_cents;

ALTER TABLE marketplace_orders
    ADD CONSTRAINT marketplace_orders_amount_cents_check CHECK (amount_cents >= 0);

ALTER TABLE facebook_page_listings
    DROP CONSTRAINT IF EXISTS facebook_page_listings_price_usd_check,
    DROP CONSTRAINT IF EXISTS facebook_page_listings_currency_usd_check;

ALTER TABLE facebook_page_listings
    ALTER COLUMN price_usd TYPE BIGINT
    USING ROUND(price_usd * 100)::bigint;

ALTER TABLE facebook_page_listings
    RENAME COLUMN price_usd TO price_cents;

ALTER TABLE facebook_page_listings
    ADD CONSTRAINT facebook_page_listings_price_cents_check CHECK (price_cents >= 0);
