package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	secret := []byte(must("SESSION_BRIDGE_SECRET"))
	outpost := must("OUTPOST_URL")
	store, err := NewStore("/data/bridge.db")
	if err != nil {
		log.Fatal(err)
	}

	signer := NewSigner(secret)
	authz := NewAuthzHandler(signer, outpost)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	mux.Handle("/authz", authz)
	mux.Handle("/transfer/new", NewTransferNewHandler(store, authz))

	claim := NewTransferClaimHandler(store, signer)
	rl := NewRateLimiter(5, 12*time.Second)
	mux.Handle("/transfer/claim", rl.Middleware(claim))

	log.Println("session-bridge listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", logRequests(mux)))
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sr := &statusRecorder{ResponseWriter: w, status: 200}
		start := time.Now()
		next.ServeHTTP(sr, r)
		user := w.Header().Get("X-Authentik-Username")
		if user == "" {
			user = "-"
		}
		log.Printf("%s %s -> %d user=%s dur=%s",
			r.Method, r.URL.Path, sr.status, user, time.Since(start))
	})
}

func must(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("env %s required", k)
	}
	return v
}
