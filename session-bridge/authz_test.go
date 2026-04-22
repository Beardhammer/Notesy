package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthzValidTransferCookie(t *testing.T) {
	s := NewSigner([]byte("k"))
	tok, _ := s.Sign("alice", time.Hour)
	h := NewAuthzHandler(s, "http://unused")

	r := httptest.NewRequest("GET", "/authz", nil)
	r.AddCookie(&http.Cookie{Name: "notesy-transfer", Value: tok})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Fatalf("code=%d", w.Code)
	}
	if got := w.Header().Get("X-Authentik-Username"); got != "alice" {
		t.Fatalf("header=%q", got)
	}
}

// oauth2-proxy's /oauth2/auth endpoint returns 2xx with X-Auth-Request-User
// on success. session-bridge renames it to X-Authentik-Username so downstream
// apps see a stable header name regardless of which sidecar is in use.
func TestAuthzNoCookieRenamesOAuth2ProxyHeader(t *testing.T) {
	called := false
	var calledPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		calledPath = r.URL.Path
		w.Header().Set("X-Auth-Request-User", "from-oauth2-proxy")
		w.Header().Set("X-Auth-Request-Email", "from-oauth2-proxy@example.com")
		w.WriteHeader(202)
	}))
	defer upstream.Close()

	h := NewAuthzHandler(NewSigner([]byte("k")), upstream.URL)
	r := httptest.NewRequest("GET", "/authz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if !called {
		t.Fatal("upstream not called")
	}
	if calledPath != "/oauth2/auth" {
		t.Fatalf("upstream called at %q; want /oauth2/auth", calledPath)
	}
	if w.Code != 200 {
		t.Fatalf("status=%d; want 200", w.Code)
	}
	if got := w.Header().Get("X-Authentik-Username"); got != "from-oauth2-proxy" {
		t.Fatalf("X-Authentik-Username = %q; want from-oauth2-proxy", got)
	}
	if got := w.Header().Get("X-Authentik-Email"); got != "from-oauth2-proxy@example.com" {
		t.Fatalf("X-Authentik-Email = %q", got)
	}
}

// When oauth2-proxy returns 401 (unauthenticated), session-bridge returns 401
// so Caddy's handle_response can redirect to /oauth2/start.
func TestAuthzReturns401WhenUnauthenticated(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(401)
	}))
	defer upstream.Close()

	h := NewAuthzHandler(NewSigner([]byte("k")), upstream.URL)
	r := httptest.NewRequest("GET", "/authz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != 401 {
		t.Fatalf("code=%d; want 401", w.Code)
	}
}

// A client-supplied X-Authentik-Username on the authz request must be
// ignored — authentication comes solely from the transfer cookie or
// oauth2-proxy's own cookie.
func TestAuthzIgnoresForgedHeader(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(401) // oauth2-proxy says no
	}))
	defer upstream.Close()

	h := NewAuthzHandler(NewSigner([]byte("k")), upstream.URL)
	r := httptest.NewRequest("GET", "/authz", nil)
	r.Header.Set("X-Authentik-Username", "evil")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != 401 {
		t.Fatalf("code=%d; want 401 (forged header must not grant auth)", w.Code)
	}
	if got := w.Header().Get("X-Authentik-Username"); got == "evil" {
		t.Fatal("forged username leaked into response")
	}
}
