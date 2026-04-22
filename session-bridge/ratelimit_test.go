package main

import (
	"testing"
	"time"
)

func TestRateLimiterCapacity(t *testing.T) {
	rl := NewRateLimiter(3, time.Hour)
	for i := 0; i < 3; i++ {
		if !rl.Allow("1.2.3.4") {
			t.Fatalf("allow %d should pass", i)
		}
	}
	if rl.Allow("1.2.3.4") {
		t.Fatal("4th should be denied")
	}
}

func TestRateLimiterRefill(t *testing.T) {
	rl := NewRateLimiter(1, 10*time.Millisecond)
	if !rl.Allow("x") {
		t.Fatal("first allowed")
	}
	if rl.Allow("x") {
		t.Fatal("second should deny")
	}
	time.Sleep(15 * time.Millisecond)
	if !rl.Allow("x") {
		t.Fatal("after refill, should allow")
	}
}

func TestRateLimiterPerKey(t *testing.T) {
	rl := NewRateLimiter(1, time.Hour)
	if !rl.Allow("a") {
		t.Fatal("a first allow")
	}
	if !rl.Allow("b") {
		t.Fatal("b independent allow")
	}
}
