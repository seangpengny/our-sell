-- +goose Up

CREATE TABLE facebook_pages (
    id UUID PRIMARY KEY,
    facebook_page_id TEXT NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(255),
    picture_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX facebook_pages_name_idx ON facebook_pages (name);

CREATE TABLE facebook_page_connections (
    page_id UUID NOT NULL REFERENCES facebook_pages(id) ON DELETE CASCADE,
    connection_id UUID NOT NULL REFERENCES facebook_connections(id) ON DELETE CASCADE,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (page_id, connection_id)
);

CREATE INDEX facebook_page_connections_connection_idx
    ON facebook_page_connections (connection_id, last_seen_at DESC);

CREATE TABLE facebook_page_listings (
    id UUID PRIMARY KEY,
    page_id UUID NOT NULL UNIQUE REFERENCES facebook_pages(id) ON DELETE RESTRICT,
    seller_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    subtitle VARCHAR(255) NOT NULL DEFAULT 'Facebook Page',
    description TEXT NOT NULL DEFAULT '',
    price_cents BIGINT NOT NULL DEFAULT 0 CHECK (price_cents >= 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    delivery_window VARCHAR(120) NOT NULL DEFAULT 'Same-day handover',
    status VARCHAR(30) NOT NULL DEFAULT 'draft',
    featured BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ,
    CONSTRAINT facebook_page_listings_status_check CHECK (
        status IN ('draft', 'pending_review', 'published', 'reserved', 'sold', 'transfer_pending', 'transferred', 'cancelled', 'unlisted')
    )
);

CREATE INDEX facebook_page_listings_status_idx
    ON facebook_page_listings (status, featured DESC, sort_order, published_at DESC);
CREATE INDEX facebook_page_listings_seller_idx
    ON facebook_page_listings (seller_user_id, updated_at DESC);

-- +goose Down

DROP TABLE IF EXISTS facebook_page_listings;
DROP TABLE IF EXISTS facebook_page_connections;
DROP TABLE IF EXISTS facebook_pages;
