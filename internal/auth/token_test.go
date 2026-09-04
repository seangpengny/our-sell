package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAccessTokenValidation(t *testing.T) {
	service, err := NewTokenService("01234567890123456789012345678901", "our-sell-api", "our-sell-client", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.MustParse("0198c8d4-6f44-7c2d-9e4b-0c4f8a4d7d3e")
	sessionID := uuid.MustParse("0198c8d4-6f44-7c2e-9e4b-0c4f8a4d7d3e")
	now := time.Now().UTC()
	raw, err := service.GenerateAccessToken(userID, sessionID, "user", now)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := service.ValidateAccessToken(raw)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}
	if claims.Subject != userID.String() || claims.SessionID != sessionID.String() || claims.Role != "user" {
		t.Fatalf("unexpected claims: %+v", claims)
	}

	otherIssuer, _ := NewTokenService("01234567890123456789012345678901", "other-issuer", "our-sell-client", 15*time.Minute)
	wrongIssuer, _ := otherIssuer.GenerateAccessToken(userID, sessionID, "user", now)
	if _, err := service.ValidateAccessToken(wrongIssuer); err == nil {
		t.Fatal("wrong issuer token was accepted")
	}
	otherAudience, _ := NewTokenService("01234567890123456789012345678901", "our-sell-api", "other-client", 15*time.Minute)
	wrongAudience, _ := otherAudience.GenerateAccessToken(userID, sessionID, "user", now)
	if _, err := service.ValidateAccessToken(wrongAudience); err == nil {
		t.Fatal("wrong audience token was accepted")
	}
	otherSecret, _ := NewTokenService("abcdef0123456789abcdef0123456789", "our-sell-api", "our-sell-client", 15*time.Minute)
	wrongSignature, _ := otherSecret.GenerateAccessToken(userID, sessionID, "user", now)
	if _, err := service.ValidateAccessToken(wrongSignature); err == nil {
		t.Fatal("invalid signature token was accepted")
	}

	short, _ := NewTokenService("01234567890123456789012345678901", "our-sell-api", "our-sell-client", time.Minute)
	expired, _ := short.GenerateAccessToken(userID, sessionID, "user", now.Add(-2*time.Minute))
	if _, err := short.ValidateAccessToken(expired); err == nil {
		t.Fatal("expired token was accepted")
	}
}
