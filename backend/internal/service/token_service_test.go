package service

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTTokenService_GenerateAndVerify(t *testing.T) {
	// Create a new token service with a test secret and expiration time
	tokenTest := NewTokenService([]byte("test-secret"), 1*time.Hour)

	// Generate a token for a test user ID
	gen, err := tokenTest.Generate("test-user-id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if gen == "" {
		t.Fatal("expected a token, got empty string")
	}

	// Verify the generated token
	verify, err := tokenTest.Verify(gen)
	if err != nil {
		t.Fatalf("expected no error on verify, got %v", err)
	}

	if verify.UserID != "test-user-id" {
		t.Fatalf("expected userID 'test-user-id', got %s", verify.UserID)
	}

	// Check that the expiration duration is 1 hour
	duration := verify.ExpiresAt.Time.Sub(verify.IssuedAt.Time)

	if duration != time.Hour {
		t.Fatalf("expected expiration duration 1 hour, got %v", duration)
	}

	// Check that the token is not expired and was issued in the past
	if time.Until(verify.ExpiresAt.Time) <= 0 {
		t.Fatal("expected token to be valid, but it is expired")
	}
	if time.Until(verify.IssuedAt.Time) > 0 {
		t.Fatal("expected token to be issued in the past, but it is in the future")
	}
}

func TestVerifyToken_Expired(t *testing.T) {
	// Create a new token service with a test secret and expiration time of 1 second
	tokenTest := NewTokenService([]byte("test-secret"), 1*time.Second)

	// Generate a token for a test user ID
	gen, err := tokenTest.Generate("test-user-id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Wait for 2 seconds to ensure the token is expired
	time.Sleep(2 * time.Second)

	// Verify the generated token
	_, err = tokenTest.Verify(gen)
	if err == nil {
		t.Fatal("expected error on verify for expired token, got nil")
	}
	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("expected jwt.ErrTokenExpired, got %v", err)
	}
}

func TestVerifyToken_InvalidSignature(t *testing.T) {	
	// Create a new token service with a test secret and expiration time
	tokenTest := NewTokenService([]byte("test-secret"), 1*time.Hour)

	// Generate a token for a test user ID
	gen, err := tokenTest.Generate("test-user-id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	
	// Create a new token service with a different secret
	tokenTestInvalid := NewTokenService([]byte("invalid-secret"), 1*time.Hour)

	// Verify the generated token with the invalid secret
	_, err = tokenTestInvalid.Verify(gen)
	if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		t.Fatalf("expected jwt.ErrTokenSignatureInvalid, got %v", err)
	}
}

func TestVerifyToken_InvalidFormat(t *testing.T){
	// Create a new token service with a test secret and expiration time
	tokenTest := NewTokenService([]byte("test-secret"), 1*time.Hour)

	// Verify an invalid token format
	_, err := tokenTest.Verify("invalid-token-format")
	if !errors.Is(err, jwt.ErrTokenMalformed) {
		t.Fatalf("expected jwt.ErrTokenMalformed, got %v", err)
	}
}

func TestVerifyToken_UnexpectedSigningMethod(t *testing.T) {
	// Create a token service that only accepts HS256
	tokenTest := NewTokenService([]byte("test-secret"), 1*time.Hour)

	// Create a token with an unexpected signing method (e.g., HS384)
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, &Claims{
		UserID: "test-user-id",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	})

	// Sign the token with the test secret using HS384
	testSecret := []byte("test-secret")
	gen, err := token.SignedString(testSecret)
	if err != nil {
		t.Fatalf("expected no error signing token, got %v", err)
	}

	// Verify the generated token with the JWTTokenService
	_, err = tokenTest.Verify(gen)
	if err == nil {
		t.Fatal("expected error for unexpected signing method, got nil")
	}
}