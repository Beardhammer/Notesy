package http

import (
	"crypto/tls"
	"net/http"
	"time"
)

var (
	// Create a custom TLS config
	tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	// Create a custom HTTP transport
	transport = &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Create a global custom HTTP client using the transport
	CustomHttpClient = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
)