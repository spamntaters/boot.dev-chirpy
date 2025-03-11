package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	secret := "bar"
	expiresIn := time.Duration(60 * 10000)
	_, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("failed to get jwt: %v", err)
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "bar"
	expiresIn := 5 * time.Minute
	jwt, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("Failed to get jwt: %v", err)
	}
	id, err := ValidateJWT(jwt, secret)
	if err != nil {
		t.Fatalf("Failed to run validation of token %v: %v", jwt, err)
	}
	if id != userID {
		t.Fatalf("User id doesn't match claims")
	}
}
