package facebook

import (
	"time"

	"github.com/google/uuid"
)

const (
	requestedScope       = "pages_show_list"
	oauthStateTTL        = 10 * time.Minute
	reservationTTL       = 15 * time.Minute
	maxPageRequests      = 100
	maxMetaResponseBytes = 1 << 20
)

type Connection struct {
	ID             string     `json:"id"`
	FacebookUserID string     `json:"facebook_account_id"`
	Name           string     `json:"name"`
	ConnectedAt    time.Time  `json:"connected_at"`
	TokenExpiresAt *time.Time `json:"token_expires_at,omitempty"`
}

type Page struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PageInventoryItem struct {
	ID             string     `json:"id"`
	FacebookPageID string     `json:"facebook_page_id"`
	Name           string     `json:"name"`
	ConnectionID   string     `json:"connection_id"`
	ConnectionName string     `json:"connection_name"`
	LastSyncedAt   *time.Time `json:"last_synced_at,omitempty"`
	Listing        *Listing   `json:"listing,omitempty"`
}

type PageInventory struct {
	Pages      []PageInventoryItem `json:"pages"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

type Listing struct {
	ID             string     `json:"id"`
	PageID         string     `json:"page_id"`
	Title          string     `json:"title"`
	Subtitle       string     `json:"subtitle"`
	Description    string     `json:"description"`
	PriceUSD       float64    `json:"price_usd"`
	Currency       string     `json:"currency"`
	DeliveryWindow string     `json:"delivery_window"`
	Status         string     `json:"status"`
	Featured       bool       `json:"featured"`
	SortOrder      int        `json:"sort_order"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
}

type ListingInput struct {
	Title          string  `json:"title"`
	Subtitle       string  `json:"subtitle"`
	Description    string  `json:"description"`
	PriceUSD       float64 `json:"price_usd"`
	Currency       string  `json:"currency"`
	DeliveryWindow string  `json:"delivery_window"`
	Status         string  `json:"status"`
	Featured       bool    `json:"featured"`
	SortOrder      int     `json:"sort_order"`
}

type PublicListing struct {
	Listing
	FacebookPageID string `json:"facebook_page_id"`
	PageName       string `json:"page_name"`
}

type MarketplaceOrder struct {
	ID                   string     `json:"id"`
	Reference            string     `json:"reference"`
	ListingID            string     `json:"listing_id"`
	BuyerUserID          string     `json:"buyer_user_id,omitempty"`
	BuyerNote            string     `json:"buyer_note,omitempty"`
	AmountUSD            float64    `json:"amount_usd"`
	Currency             string     `json:"currency"`
	PaymentMethod        string     `json:"payment_method,omitempty"`
	Status               string     `json:"status"`
	ReservationExpiresAt time.Time  `json:"reservation_expires_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	PaidAt               *time.Time `json:"paid_at,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
}

type MarketplaceOrderInput struct {
	BuyerNote     string `json:"buyer_note"`
	PaymentMethod string `json:"payment_method"`
	TermsAccepted bool   `json:"terms_accepted"`
}

type pageInventoryParams struct {
	UserID       uuid.UUID
	ConnectionID uuid.UUID
	Search       string
	Page         int
	PageSize     int
}

type storedConnection struct {
	Connection
	UserID               string
	EncryptedAccessToken []byte
	Scopes               string
}

type oauthState struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
}

type CallbackParams struct {
	Code        string
	State       string
	OAuthError  string
	ErrorReason string
}

type CallbackStatus string

const (
	CallbackConnected CallbackStatus = "connected"
	CallbackCancelled CallbackStatus = "cancelled"
	CallbackError     CallbackStatus = "error"
	CallbackStateErr  CallbackStatus = "invalid_state"
)
