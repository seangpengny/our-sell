-- name: UpsertFacebookPage :one
INSERT INTO facebook_pages (
    id, facebook_page_id, name, last_synced_at, is_active
)
VALUES ($1, $2, $3, NOW(), TRUE)
ON CONFLICT (facebook_page_id) DO UPDATE
SET name = EXCLUDED.name,
    is_active = TRUE,
    last_synced_at = NOW(),
    updated_at = NOW()
RETURNING id, facebook_page_id, name, category, picture_url, is_active,
          last_synced_at, created_at, updated_at;

-- name: LinkFacebookPageToConnection :exec
INSERT INTO facebook_page_connections (page_id, connection_id, last_seen_at)
VALUES ($1, $2, NOW())
ON CONFLICT (page_id, connection_id) DO UPDATE
SET last_seen_at = NOW();

-- name: ListFacebookPageInventory :many
SELECT DISTINCT ON (p.id)
    p.id, p.facebook_page_id, p.name, p.category, p.picture_url, p.is_active,
    p.last_synced_at, p.created_at, p.updated_at,
    c.id AS connection_id, c.facebook_user_name AS connection_name,
    l.id AS listing_id, l.title AS listing_title, l.subtitle AS listing_subtitle,
    l.description AS listing_description, COALESCE(l.price_usd, 0)::double precision AS price_usd, l.currency,
    l.delivery_window, l.status AS listing_status, l.featured, l.sort_order,
    l.published_at
FROM facebook_pages p
JOIN facebook_page_connections pc ON pc.page_id = p.id
JOIN facebook_connections c ON c.id = pc.connection_id
    AND c.user_id = $1 AND c.revoked_at IS NULL
LEFT JOIN facebook_page_listings l ON l.page_id = p.id AND l.seller_user_id = $1
WHERE ($2 = '00000000-0000-0000-0000-000000000000'::uuid OR c.id = $2)
  AND ($3::text = '' OR p.name ILIKE '%' || $3::text || '%' OR p.facebook_page_id ILIKE '%' || $3::text || '%')
ORDER BY p.id, pc.last_seen_at DESC
LIMIT $4 OFFSET $5;

-- name: CountFacebookPageInventory :one
SELECT COUNT(DISTINCT p.id)
FROM facebook_pages p
JOIN facebook_page_connections pc ON pc.page_id = p.id
JOIN facebook_connections c ON c.id = pc.connection_id
    AND c.user_id = $1 AND c.revoked_at IS NULL
WHERE ($2 = '00000000-0000-0000-0000-000000000000'::uuid OR c.id = $2)
  AND ($3::text = '' OR p.name ILIKE '%' || $3::text || '%' OR p.facebook_page_id ILIKE '%' || $3::text || '%');

-- name: GetFacebookPageForUser :one
SELECT p.id, p.facebook_page_id, p.name, p.category, p.picture_url, p.is_active,
       p.last_synced_at, p.created_at, p.updated_at,
       c.id AS connection_id, c.user_id AS connection_user_id
FROM facebook_pages p
JOIN facebook_page_connections pc ON pc.page_id = p.id
JOIN facebook_connections c ON c.id = pc.connection_id
    AND c.user_id = $2 AND c.revoked_at IS NULL
WHERE p.id = $1
ORDER BY pc.last_seen_at DESC
LIMIT 1;

-- name: UpsertFacebookPageListing :one
INSERT INTO facebook_page_listings (
    id, page_id, seller_user_id, title, subtitle, description, price_usd,
    currency, delivery_window, status, featured, sort_order, published_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7::double precision, $8, $9, $10, $11, $12,
        CASE WHEN $10::varchar = 'published' THEN NOW() ELSE NULL END)
ON CONFLICT (page_id) DO UPDATE
SET title = EXCLUDED.title,
    subtitle = EXCLUDED.subtitle,
    description = EXCLUDED.description,
    price_usd = EXCLUDED.price_usd,
    currency = EXCLUDED.currency,
    delivery_window = EXCLUDED.delivery_window,
    status = EXCLUDED.status,
    featured = EXCLUDED.featured,
    sort_order = EXCLUDED.sort_order,
    published_at = CASE
        WHEN EXCLUDED.status = 'published' AND facebook_page_listings.published_at IS NULL THEN NOW()
        WHEN EXCLUDED.status <> 'published' THEN NULL
        ELSE facebook_page_listings.published_at
    END,
    updated_at = NOW()
WHERE facebook_page_listings.seller_user_id = EXCLUDED.seller_user_id
RETURNING id, page_id, seller_user_id, title, subtitle, description, price_usd::double precision AS price_usd,
          currency, delivery_window, status, featured, sort_order, created_at,
          updated_at, published_at;

-- name: ListPublishedFacebookPageListings :many
SELECT l.id, l.page_id, p.facebook_page_id, p.name, l.title, l.subtitle,
       l.description, l.price_usd::double precision AS price_usd, l.currency, l.delivery_window, l.status,
       l.featured, l.sort_order, l.created_at, l.updated_at, l.published_at
FROM facebook_page_listings l
JOIN facebook_pages p ON p.id = l.page_id AND p.is_active = TRUE
WHERE l.status = 'published'
  AND ($1::text = '' OR l.title ILIKE '%' || $1::text || '%' OR p.name ILIKE '%' || $1::text || '%'
       OR p.facebook_page_id ILIKE '%' || $1::text || '%')
  AND EXISTS (
      SELECT 1
      FROM facebook_page_connections pc
      JOIN facebook_connections c ON c.id = pc.connection_id
      WHERE pc.page_id = l.page_id
        AND c.user_id = l.seller_user_id
        AND c.revoked_at IS NULL
  )
ORDER BY l.featured DESC, l.sort_order, l.published_at DESC, l.id DESC
LIMIT $2 OFFSET $3;

-- name: CountPublishedFacebookPageListings :one
SELECT COUNT(*)
FROM facebook_page_listings l
JOIN facebook_pages p ON p.id = l.page_id AND p.is_active = TRUE
WHERE l.status = 'published'
  AND ($1::text = '' OR l.title ILIKE '%' || $1::text || '%' OR p.name ILIKE '%' || $1::text || '%'
       OR p.facebook_page_id ILIKE '%' || $1::text || '%')
  AND EXISTS (
      SELECT 1
      FROM facebook_page_connections pc
      JOIN facebook_connections c ON c.id = pc.connection_id
      WHERE pc.page_id = l.page_id
        AND c.user_id = l.seller_user_id
        AND c.revoked_at IS NULL
  );

-- name: GetPublishedFacebookPageListing :one
SELECT l.id, l.page_id, p.facebook_page_id, p.name, l.title, l.subtitle,
       l.description, l.price_usd::double precision AS price_usd, l.currency, l.delivery_window, l.status,
       l.featured, l.sort_order, l.created_at, l.updated_at, l.published_at
FROM facebook_page_listings l
JOIN facebook_pages p ON p.id = l.page_id AND p.is_active = TRUE
WHERE l.id = $1 AND l.status = 'published'
  AND EXISTS (
      SELECT 1
      FROM facebook_page_connections pc
      JOIN facebook_connections c ON c.id = pc.connection_id
      WHERE pc.page_id = l.page_id
        AND c.user_id = l.seller_user_id
        AND c.revoked_at IS NULL
  );
