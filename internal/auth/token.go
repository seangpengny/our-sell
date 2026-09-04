package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	SessionID string `json:"session_id"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

type TokenService struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
}

func NewTokenService(secret, issuer, audience string, ttl time.Duration) (*TokenService, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT secret must be at least 32 bytes")
	}
	if issuer == "" || audience == "" || ttl <= 0 {
		return nil, errors.New("JWT issuer, audience, and TTL are required")
	}
	return &TokenService{secret: []byte(secret), issuer: issuer, audience: audience, ttl: ttl}, nil
}

func (s *TokenService) GenerateAccessToken(userID, sessionID uuid.UUID, role string, now time.Time) (string, error) {
	now = now.UTC()
	claims := Claims{
		SessionID: sessionID.String(),
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID.String(),
			Audience:  jwt.ClaimStrings{s.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *TokenService) ValidateAccessToken(raw string) (Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(raw, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected JWT signing method %q", token.Method.Alg())
		}
		return s.secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(5*time.Second),
	)
	if err != nil || token == nil || !token.Valid {
		return Claims{}, errors.New("invalid access token")
	}
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return Claims{}, errors.New("access token subject is invalid")
	}
	if _, err := uuid.Parse(claims.SessionID); err != nil {
		return Claims{}, errors.New("access token session is invalid")
	}
	if claims.Role == "" {
		return Claims{}, errors.New("access token role is missing")
	}
	return claims, nil
}
