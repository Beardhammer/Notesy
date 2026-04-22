package main

import (
	"io"
	"net/http"
	"net/url"
)

type AuthzHandler struct {
	signer  *Signer
	outpost *url.URL
	client  *http.Client
}

func NewAuthzHandler(s *Signer, outpostURL string) *AuthzHandler {
	u, _ := url.Parse(outpostURL)
	return &AuthzHandler{signer: s, outpost: u, client: &http.Client{}}
}

// authResult is the identity derived from an authenticated request.
type authResult struct {
	Username string
	Email    string
	Groups   string
}

// Authenticate validates the request independently of any client-supplied
// X-Authentik-* headers. It checks the notesy-transfer HMAC cookie first,
// then falls back to oauth2-proxy's /oauth2/auth endpoint. Returns nil when
// the request is unauthenticated.
func (h *AuthzHandler) Authenticate(r *http.Request) (*authResult, error) {
	if c, err := r.Cookie("notesy-transfer"); err == nil {
		if sub, err := h.signer.Verify(c.Value); err == nil {
			return &authResult{Username: sub}, nil
		}
	}

	target := *h.outpost
	target.Path = "/oauth2/auth"
	req, _ := http.NewRequestWithContext(r.Context(), http.MethodGet, target.String(), nil)
	if cookie := r.Header.Get("Cookie"); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil
	}
	user := resp.Header.Get("X-Auth-Request-User")
	if user == "" {
		return nil, nil
	}
	return &authResult{
		Username: user,
		Email:    resp.Header.Get("X-Auth-Request-Email"),
		Groups:   resp.Header.Get("X-Auth-Request-Groups"),
	}, nil
}

func (h *AuthzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.Authenticate(r)
	if err != nil {
		http.Error(w, "upstream auth unreachable", http.StatusBadGateway)
		return
	}
	if res == nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	w.Header().Set("X-Authentik-Username", res.Username)
	if res.Email != "" {
		w.Header().Set("X-Authentik-Email", res.Email)
	}
	if res.Groups != "" {
		w.Header().Set("X-Authentik-Groups", res.Groups)
	}
	w.WriteHeader(http.StatusOK)
}
