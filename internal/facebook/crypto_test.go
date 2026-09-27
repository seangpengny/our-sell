package facebook

import (
	"bytes"
	"testing"
)

func TestTokenCipherRoundTripAndTamperDetection(t *testing.T) {
	cipher, err := NewTokenCipher(bytes.Repeat([]byte{7}, 32))
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := cipher.encrypt("facebook-access-token")
	if err != nil {
		t.Fatal(err)
	}
	if string(ciphertext) == "facebook-access-token" {
		t.Fatal("token was not encrypted")
	}
	plaintext, err := cipher.decrypt(ciphertext)
	if err != nil || plaintext != "facebook-access-token" {
		t.Fatalf("decrypt() = %q, %v", plaintext, err)
	}
	ciphertext[len(ciphertext)-1] ^= 1
	if _, err := cipher.decrypt(ciphertext); err == nil {
		t.Fatal("tampered token decrypted successfully")
	}
}

func TestOAuthStateIsUnpredictableAndKeyDoesNotContainRawState(t *testing.T) {
	first, err := newOAuthState()
	if err != nil {
		t.Fatal(err)
	}
	second, err := newOAuthState()
	if err != nil {
		t.Fatal(err)
	}
	if first == second || len(first) != 64 || stateKey(first) == "facebook:oauth:state:"+first {
		t.Fatalf("OAuth state is not safely generated/keyed: %q", first)
	}
}
