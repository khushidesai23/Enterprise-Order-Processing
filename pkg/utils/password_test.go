package utils

import "testing"

func TestPasswordHelpers(t *testing.T) {
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "password123" {
		t.Fatal("expected hashed password to differ from input")
	}

	if !CheckPassword(hash, "password123") {
		t.Fatal("expected password to match generated hash")
	}

	if CheckPassword(hash, "wrong-password") {
		t.Fatal("expected wrong password to fail")
	}
}
