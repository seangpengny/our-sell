package response

import "testing"

func TestDecodeJSONRejectsUnknownAndTrailingValues(t *testing.T) {
	var destination struct {
		Email string `json:"email"`
	}
	if err := DecodeJSON([]byte(`{"email":"user@example.com","extra":true}`), &destination); err == nil {
		t.Fatal("unknown field was accepted")
	}
	if err := DecodeJSON([]byte(`{"email":"user@example.com"}{"second":true}`), &destination); err == nil {
		t.Fatal("trailing JSON value was accepted")
	}
}
