-- +goose Up

ALTER TABLE marketplace_orders
    ADD COLUMN buyer_note TEXT NOT NULL DEFAULT '';

-- Preserve details from pre-login orders before removing the duplicated columns.
UPDATE marketplace_orders
SET buyer_note = concat_ws(
        E'\n',
        NULLIF('Name: ' || buyer_name, 'Name: '),
        NULLIF('Email: ' || buyer_email, 'Email: '),
        NULLIF('Telegram: ' || buyer_telegram, 'Telegram: ')
    )
WHERE buyer_user_id IS NULL
  AND buyer_note = '';

DROP INDEX IF EXISTS marketplace_orders_buyer_email_idx;

ALTER TABLE marketplace_orders
    DROP COLUMN buyer_name,
    DROP COLUMN buyer_email,
    DROP COLUMN buyer_telegram;

-- Existing legacy orders may not have a user. New rows are protected by this
-- constraint even though it remains NOT VALID until legacy rows are resolved.
ALTER TABLE marketplace_orders
    ADD CONSTRAINT marketplace_orders_buyer_user_id_required
    CHECK (buyer_user_id IS NOT NULL) NOT VALID;

-- A buyer must remain linked to an order once an order is created.
ALTER TABLE marketplace_orders
    DROP CONSTRAINT IF EXISTS marketplace_orders_buyer_user_id_fkey;

ALTER TABLE marketplace_orders
    ADD CONSTRAINT marketplace_orders_buyer_user_id_fkey
    FOREIGN KEY (buyer_user_id) REFERENCES users(id) ON DELETE RESTRICT;

CREATE INDEX marketplace_orders_buyer_user_idx
    ON marketplace_orders (buyer_user_id, created_at DESC);

-- +goose Down

DROP INDEX IF EXISTS marketplace_orders_buyer_user_idx;

ALTER TABLE marketplace_orders
    DROP CONSTRAINT IF EXISTS marketplace_orders_buyer_user_id_fkey,
    DROP CONSTRAINT IF EXISTS marketplace_orders_buyer_user_id_required;

ALTER TABLE marketplace_orders
    ADD COLUMN buyer_name VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN buyer_email TEXT NOT NULL DEFAULT '',
    ADD COLUMN buyer_telegram VARCHAR(100) NOT NULL DEFAULT '';

DROP INDEX IF EXISTS marketplace_orders_buyer_email_idx;
CREATE INDEX marketplace_orders_buyer_email_idx
    ON marketplace_orders (buyer_email, created_at DESC);

ALTER TABLE marketplace_orders
    DROP COLUMN buyer_note;

ALTER TABLE marketplace_orders
    ADD CONSTRAINT marketplace_orders_buyer_user_id_fkey
    FOREIGN KEY (buyer_user_id) REFERENCES users(id) ON DELETE SET NULL;
