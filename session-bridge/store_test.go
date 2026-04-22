package main

import (
	"testing"
	"time"
)

func TestStoreIssueAndClaim(t *testing.T) {
	s, err := NewStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	code, err := s.Issue("jdoe", 10*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 8 {
		t.Fatalf("code len = %d; want 8", len(code))
	}

	sub, err := s.Claim(code)
	if err != nil {
		t.Fatal(err)
	}
	if sub != "jdoe" {
		t.Fatalf("sub = %q; want jdoe", sub)
	}
}

func TestStoreClaimReplay(t *testing.T) {
	s, _ := NewStore(":memory:")
	code, _ := s.Issue("jdoe", time.Hour)
	_, _ = s.Claim(code)
	_, err := s.Claim(code)
	if err == nil {
		t.Fatal("expected replay error")
	}
}

func TestStoreClaimExpired(t *testing.T) {
	s, _ := NewStore(":memory:")
	code, _ := s.Issue("jdoe", -time.Second)
	_, err := s.Claim(code)
	if err == nil {
		t.Fatal("expected expired error")
	}
}

func TestStoreClaimUnknown(t *testing.T) {
	s, _ := NewStore(":memory:")
	_, err := s.Claim("00000000")
	if err == nil {
		t.Fatal("expected unknown-code error")
	}
}
