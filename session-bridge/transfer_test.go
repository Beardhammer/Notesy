package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// testAuthz returns a TransferNewHandler whose AuthzHandler points at a stub
// upstream that validates cookies/headers the way oauth2-proxy would for
// testing purposes.
func testNewHandler(t *testing.T, signer *Signer) (*TransferNewHandler, *httptest.Server) {
	t.Helper()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Cookie") == "oauth2_proxy_stub=valid" {
			w.Header().Set("X-Auth-Request-User", "jdoe-from-oauth2")
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	store, _ := NewStore(":memory:")
	authz := NewAuthzHandler(signer, upstream.URL)
	return NewTransferNewHandler(store, authz), upstream
}

func TestTransferNewReturnsCode_ViaTransferCookie(t *testing.T) {
	s := NewSigner([]byte("k"))
	h, up := testNewHandler(t, s)
	defer up.Close()

	tok, _ := s.Sign("jdoe", time.Hour)
	r := httptest.NewRequest("POST", "/transfer/new", nil)
	r.Header.Set("Accept", "application/json")
	r.AddCookie(&http.Cookie{Name: "notesy-transfer", Value: tok})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Fatalf("code=%d", w.Code)
	}
	var body struct {
		Code             string `json:"code"`
		ExpiresInSeconds int    `json:"expiresInSeconds"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Code) != 8 || strings.TrimLeft(body.Code, "0123456789") != "" {
		t.Fatalf("code=%q not 8 digits", body.Code)
	}
	if body.ExpiresInSeconds != 600 {
		t.Fatalf("expiry=%d", body.ExpiresInSeconds)
	}
}

func TestTransferNewReturnsCode_ViaOAuth2Cookie(t *testing.T) {
	s := NewSigner([]byte("k"))
	h, up := testNewHandler(t, s)
	defer up.Close()

	r := httptest.NewRequest("POST", "/transfer/new", nil)
	r.Header.Set("Accept", "application/json")
	r.AddCookie(&http.Cookie{Name: "oauth2_proxy_stub", Value: "valid"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Fatalf("code=%d", w.Code)
	}
}

// Regression test for header-forgery attack: a client supplying an
// unauthenticated X-Authentik-Username must NOT get a code.
func TestTransferNewRejectsForgedHeader(t *testing.T) {
	s := NewSigner([]byte("k"))
	h, up := testNewHandler(t, s)
	defer up.Close()

	r := httptest.NewRequest("POST", "/transfer/new", nil)
	r.Header.Set("X-Authentik-Username", "admin") // forged, no valid cookie
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != 401 {
		t.Fatalf("code=%d; want 401 for unauth request with forged header", w.Code)
	}
}

func TestTransferNewRequiresAuth(t *testing.T) {
	s := NewSigner([]byte("k"))
	h, up := testNewHandler(t, s)
	defer up.Close()

	r := httptest.NewRequest("POST", "/transfer/new", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != 401 {
		t.Fatalf("code=%d; want 401", w.Code)
	}
}

func TestClaimGETShowsForm(t *testing.T) {
	h := NewTransferClaimHandler(newTestStore(t), NewSigner([]byte("k")))
	r := httptest.NewRequest("GET", "/transfer/claim", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("code=%d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "<form") {
		t.Fatal("no form in html")
	}
}

func TestClaimPOSTSetsCookieAndRedirects(t *testing.T) {
	store := newTestStore(t)
	code, _ := store.Issue("alice", time.Hour)
	h := NewTransferClaimHandler(store, NewSigner([]byte("k")))

	r := httptest.NewRequest("POST", "/transfer/claim", strings.NewReader("code="+code))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != 302 {
		t.Fatalf("code=%d; want 302", w.Code)
	}
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "notesy-transfer" {
			found = true
		}
	}
	if !found {
		t.Fatal("no notesy-transfer cookie")
	}
}

func TestClaimPOSTInvalidCode(t *testing.T) {
	h := NewTransferClaimHandler(newTestStore(t), NewSigner([]byte("k")))
	r := httptest.NewRequest("POST", "/transfer/claim",
		strings.NewReader("code=99999999"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatalf("code=%d", w.Code)
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	return s
}
