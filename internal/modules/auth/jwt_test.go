package auth

import (
	"testing"
	"time"
)

func TestJWTManagerGenerateAndVerifyToken(t *testing.T) {
	manager := NewJWTManager("secret", time.Hour)

	token, err := manager.GenerateToken("user-1", "user@example.com")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := manager.VerifyToken(token)
	if err != nil {
		t.Fatalf("VerifyToken() error = %v", err)
	}

	if claims.UserID != "user-1" {
		t.Fatalf("expected user id user-1, got %s", claims.UserID)
	}

	if claims.Email != "user@example.com" {
		t.Fatalf("expected email user@example.com, got %s", claims.Email)
	}
}

func TestJWTManagerVerifyTokenRejectsInvalidToken(t *testing.T) {
	manager := NewJWTManager("secret", time.Hour)

	if _, err := manager.VerifyToken("invalid-token"); err == nil {
		t.Fatal("expected invalid token to fail verification")
	}
}
