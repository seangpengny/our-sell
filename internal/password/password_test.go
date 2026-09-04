package password

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "correct horse battery staple" || len(hash) < 80 {
		t.Fatalf("hash does not look encoded: %q", hash)
	}
	valid, err := VerifyPassword("correct horse battery staple", hash)
	if err != nil || !valid {
		t.Fatalf("VerifyPassword(correct) = %v, %v", valid, err)
	}
	valid, err = VerifyPassword("wrong password", hash)
	if err != nil || valid {
		t.Fatalf("VerifyPassword(wrong) = %v, %v", valid, err)
	}
}

func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	if valid, err := VerifyPassword("password", "not-an-argon2-hash"); err == nil || valid {
		t.Fatalf("VerifyPassword(malformed) = %v, %v", valid, err)
	}
	if !NeedsRehash("not-an-argon2-hash") {
		t.Fatal("NeedsRehash should reject malformed hashes")
	}
}
