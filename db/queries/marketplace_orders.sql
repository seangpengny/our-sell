-- name: ReleaseExpiredMarketplaceReservations :exec
WITH expired_orders AS (
    UPDATE marketplace_orders
    SET status = 'expired', updated_at = NOW()
    WHERE (
        status = 'pending_payment'
        OR (status = 'paid' AND payment_method = 'wallet')
    )
      AND reservation_expires_at <= NOW()
    RETURNING id, listing_id, buyer_user_id, reference, amount_usd, payment_method
), wallet_release AS (
    UPDATE wallets w
    SET available_amount = w.available_amount + e.amount_usd,
        held_amount = w.held_amount - e.amount_usd,
        updated_at = NOW()
    FROM expired_orders e
    WHERE e.payment_method = 'wallet'
      AND w.user_id = e.buyer_user_id
      AND w.currency = 'USD'
      AND w.held_amount >= e.amount_usd
    RETURNING w.id AS wallet_id, w.user_id, w.available_amount,
              w.held_amount, e.reference, e.amount_usd
), ledger AS (
    INSERT INTO wallet_ledger_entries (
        id, wallet_id, user_id, entry_type, reference,
        available_delta_usd, held_delta_usd,
        available_balance_usd, held_balance_usd, description
    )
    SELECT gen_random_uuid(), wallet_id, user_id, 'order_release', reference,
           amount_usd, -amount_usd, available_amount, held_amount,
           'Held funds released after marketplace reservation expired'
    FROM wallet_release
    RETURNING id
)
UPDATE facebook_page_listings l
SET status = 'published', updated_at = NOW()
FROM expired_orders e
WHERE l.id = e.listing_id
  AND l.status = 'reserved'
RETURNING l.id;

-- name: CreateMarketplaceOrder :one
WITH candidate AS (
    SELECT l.id, l.seller_user_id, l.price_usd::double precision AS price_usd, l.currency
    FROM facebook_page_listings l
    WHERE l.id = $2
      AND l.status = 'published'
      AND EXISTS (
          SELECT 1
          FROM facebook_page_connections pc
          JOIN facebook_connections c ON c.id = pc.connection_id
          WHERE pc.page_id = l.page_id
            AND c.user_id = l.seller_user_id
            AND c.revoked_at IS NULL
      )
    FOR UPDATE
), created AS (
    INSERT INTO marketplace_orders (
        id, reference, listing_id, buyer_user_id, buyer_note, amount_usd,
        currency, payment_method, status,
        reservation_expires_at
    )
    SELECT $1, $3, id, $4, $5, price_usd, currency, $6::varchar,
           'pending_payment', $7
    FROM candidate
    RETURNING id, reference, listing_id, buyer_user_id, buyer_note, amount_usd::double precision AS amount_usd,
              currency, payment_method, status,
              reservation_expires_at, created_at, updated_at, paid_at, completed_at
)
UPDATE facebook_page_listings l
SET status = 'reserved', updated_at = NOW()
FROM created
WHERE l.id = created.listing_id
RETURNING created.id, created.reference, created.listing_id, created.buyer_user_id,
          created.buyer_note, created.amount_usd::double precision AS amount_usd, created.currency,
          created.payment_method,
          created.status, created.reservation_expires_at, created.created_at,
          created.updated_at, created.paid_at, created.completed_at;

-- name: CreateWalletMarketplaceOrder :one
WITH candidate AS (
    SELECT l.id, l.price_usd, l.currency
    FROM facebook_page_listings l
    WHERE l.id = $2
      AND l.status = 'published'
      AND EXISTS (
          SELECT 1
          FROM facebook_page_connections pc
          JOIN facebook_connections c ON c.id = pc.connection_id
          WHERE pc.page_id = l.page_id
            AND c.user_id = l.seller_user_id
            AND c.revoked_at IS NULL
      )
    FOR UPDATE
), wallet_locked AS (
    SELECT w.id AS wallet_id, w.user_id, w.available_amount, w.held_amount
    FROM wallets w
    JOIN candidate c ON TRUE
    WHERE w.user_id = $4
      AND w.currency = 'USD'
      AND w.available_amount >= c.price_usd
    FOR UPDATE
), wallet_updated AS (
    UPDATE wallets w
    SET available_amount = w.available_amount - c.price_usd,
        held_amount = w.held_amount + c.price_usd,
        updated_at = NOW()
    FROM candidate c, wallet_locked wl
    WHERE w.id = wl.wallet_id
    RETURNING w.id AS wallet_id, w.user_id, w.available_amount,
              w.held_amount, c.price_usd
), created AS (
    INSERT INTO marketplace_orders (
        id, reference, listing_id, buyer_user_id, buyer_note, amount_usd,
        currency, payment_method, status, reservation_expires_at, paid_at
    )
    SELECT $1, $3, c.id, $4, $5, c.price_usd,
           c.currency, 'wallet', 'paid', $6, NOW()
    FROM candidate c, wallet_updated wu
    RETURNING id, reference, listing_id, buyer_user_id, buyer_note,
              amount_usd::double precision AS amount_usd, currency,
              payment_method, status, reservation_expires_at, created_at,
              updated_at, paid_at, completed_at
), ledger AS (
    INSERT INTO wallet_ledger_entries (
        id, wallet_id, user_id, entry_type, reference,
        available_delta_usd, held_delta_usd,
        available_balance_usd, held_balance_usd, description
    )
    SELECT gen_random_uuid(), wu.wallet_id, wu.user_id, 'order_hold', cr.reference,
           -cr.amount_usd, cr.amount_usd, wu.available_amount, wu.held_amount,
           'Funds held for marketplace order'
    FROM created cr, wallet_updated wu
    RETURNING id
)
UPDATE facebook_page_listings l
SET status = 'reserved', updated_at = NOW()
FROM created cr, ledger
WHERE l.id = cr.listing_id
RETURNING cr.id, cr.reference, cr.listing_id, cr.buyer_user_id,
          cr.buyer_note, cr.amount_usd, cr.currency, cr.payment_method,
          cr.status, cr.reservation_expires_at, cr.created_at,
          cr.updated_at, cr.paid_at, cr.completed_at;

-- name: GetPublishedListingPrice :one
SELECT price_usd::double precision AS price_usd
FROM facebook_page_listings
WHERE id = $1 AND status = 'published';

-- name: GetMarketplaceOrder :one
SELECT id, reference, listing_id, buyer_user_id, buyer_note, amount_usd::double precision AS amount_usd,
       currency, payment_method, status,
       reservation_expires_at, created_at, updated_at, paid_at, completed_at
FROM marketplace_orders
WHERE id = $1;
