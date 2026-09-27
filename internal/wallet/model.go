package wallet

import "time"

const (
	StatusPendingPayment = "pending_payment"
	StatusConfirmed      = "confirmed"
	StatusFailed         = "failed"
	StatusExpired        = "expired"
	StatusReversed       = "reversed"
)

type Wallet struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Currency        string    `json:"currency"`
	AvailableAmount float64   `json:"available_amount"`
	HeldAmount      float64   `json:"held_amount"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type LedgerEntry struct {
	ID                  string    `json:"id"`
	WalletID            string    `json:"wallet_id"`
	UserID              string    `json:"user_id"`
	EntryType           string    `json:"entry_type"`
	Reference           string    `json:"reference"`
	AvailableDeltaUSD   float64   `json:"available_delta_usd"`
	HeldDeltaUSD        float64   `json:"held_delta_usd"`
	AvailableBalanceUSD float64   `json:"available_balance_usd"`
	HeldBalanceUSD      float64   `json:"held_balance_usd"`
	Description         string    `json:"description"`
	CreatedAt           time.Time `json:"created_at"`
}

type Topup struct {
	ID                    string     `json:"id"`
	WalletID              string     `json:"wallet_id"`
	UserID                string     `json:"user_id"`
	Reference             string     `json:"reference"`
	AmountUSD             float64    `json:"amount_usd"`
	Currency              string     `json:"currency"`
	QRPayload             string     `json:"qr_payload"`
	QRMD5                 string     `json:"qr_md5"`
	ProviderTransactionID string     `json:"provider_transaction_id,omitempty"`
	Status                string     `json:"status"`
	FailureReason         string     `json:"failure_reason,omitempty"`
	ResolutionNote        string     `json:"resolution_note,omitempty"`
	ExpiresAt             time.Time  `json:"expires_at"`
	PaidAt                *time.Time `json:"paid_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type TopupInput struct {
	AmountUSD float64 `json:"amount_usd"`
}

type LedgerResult struct {
	Entries  []LedgerEntry `json:"entries"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	HasMore  bool          `json:"has_more"`
}

type AdminTopup struct {
	Topup
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
}

type AdminTopupResult struct {
	Topups     []AdminTopup `json:"topups"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}
