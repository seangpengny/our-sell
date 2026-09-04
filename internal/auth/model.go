package auth

import "time"

type Session struct {
	ID         string     `json:"id"`
	UserID     string     `json:"-"`
	UserAgent  string     `json:"device"`
	IPAddress  string     `json:"ip"`
	ExpiresAt  time.Time  `json:"expires_at"`
	LastUsedAt time.Time  `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"-"`
}

type SessionToken struct {
	Session
	PresentedTokenUsedAt    *time.Time
	PresentedTokenRevokedAt *time.Time
}
