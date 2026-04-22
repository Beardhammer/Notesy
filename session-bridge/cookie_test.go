package main

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestSignVerifyRoundTrip(t *testing.T) {
	s := NewSigner([]byte("test-secret"))
	tok, err := s.Sign("jdoe", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	sub, err := s.Verify(tok)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if sub != "jdoe" {
		t.Fatalf("sub = %q; want jdoe", sub)
	}
}

func TestVerifyTamperedFails(t *testing.T) {
	s := NewSigner([]byte("test-secret"))
	tok, _ := s.Sign("jdoe", time.Hour)
	_, err := s.Verify(tok + "x")
	if err == nil {
		t.Fatal("expected error on tampered token")
	}
}

func TestVerifyWrongSecretFails(t *testing.T) {
	a := NewSigner([]byte("secret-a"))
	b := NewSigner([]byte("secret-b"))
	tok, _ := a.Sign("jdoe", time.Hour)
	_, err := b.Verify(tok)
	if err == nil {
		t.Fatal("expected error with wrong secret")
	}
}

func TestVerifyExpiredFails(t *testing.T) {
	s := NewSigner([]byte("test-secret"))
	tok, _ := s.Sign("jdoe", -time.Minute)
	_, err := s.Verify(tok)
	if err == nil {
		t.Fatal("expected expiry error")
	}
}

func TestVerifyRejectsNonHS256(t *testing.T) {
	// Forge a token that claims HS512 (wrong HMAC variant — still
	// rejected because WithValidMethods only allows HS256).
	s := NewSigner([]byte("test-secret"))
	claims := jwt.RegisteredClaims{
		Subject:   "mallory",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Verify(tok)
	if err == nil {
		t.Fatal("expected Verify to reject HS512 token")
	}
}
