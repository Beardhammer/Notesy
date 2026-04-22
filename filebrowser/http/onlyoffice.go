package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v4"

	"github.com/filebrowser/filebrowser/v2/files"
)

type OnlyOfficeCallback struct {
	ChangesURL string   `json:"changesurl,omitempty"`
	Key        string   `json:"key"`
	Status     int      `json:"status"`
	URL        string   `json:"url,omitempty"`
	Users      []string `json:"users,omitempty"`
	UserData   string   `json:"userdata,omitempty"`
}

var onlyofficeCallbackHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	var data OnlyOfficeCallback
	err = json.Unmarshal(body, &data)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if data.Status == 2 || data.Status == 6 {
		docPath := r.URL.Query().Get("save")
		if docPath == "" {
			return http.StatusInternalServerError, errors.New("unable to get file save path")
		}

		if !d.user.Perm.Modify || !d.Check(docPath) {
			return http.StatusForbidden, nil
		}

		// Rewrite external OnlyOffice URL to internal Docker network URL
		// OnlyOffice returns download URLs using its external domain, but
		// FileBrowser needs to reach it via the internal container hostname.
		// The configured URL may be a relative path (e.g. /oo) when using
		// path-based routing, so we reconstruct the full external origin.
		downloadURL := data.URL
		ooExternalURL := d.settings.OnlyOffice.URL
		if ooExternalURL != "" {
			if strings.HasPrefix(ooExternalURL, "/") {
				// Path-based: build full URL from request Host header
				scheme := "http"
				if r.TLS != nil {
					scheme = "https"
				}
				ooExternalURL = scheme + "://" + r.Host + ooExternalURL
			}
			downloadURL = strings.Replace(downloadURL, ooExternalURL, "http://onlyoffice", 1)
		}

		doc, err := CustomHttpClient.Get(downloadURL)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to download document: %w", err)
		}
		defer doc.Body.Close()

		if doc.StatusCode != http.StatusOK {
			return http.StatusInternalServerError, fmt.Errorf("OnlyOffice returned status %d", doc.StatusCode)
		}

		// Buffer the entire document before writing to avoid truncating the file on download failure
		content, err := io.ReadAll(doc.Body)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to read document body: %w", err)
		}

		if len(content) == 0 {
			return http.StatusInternalServerError, errors.New("OnlyOffice returned empty document, refusing to overwrite")
		}

		err = d.RunHook(func() error {
			_, writeErr := writeFile(d.user.Fs, docPath, bytes.NewReader(content))
			if writeErr != nil {
				return writeErr
			}
			return nil
		}, "save", docPath, "", d.user)

		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	resp := map[string]int{
		"error": 0,
	}
	return renderJSON(w, r, resp)
})

// onlyofficeTokenHandler signs a config payload as a JWT for the
// OnlyOffice Document Server.  This avoids using the Web Crypto API
// (crypto.subtle) on the client, which is unavailable over plain HTTP.
var onlyofficeTokenHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	secret := d.settings.OnlyOffice.JWTSecret
	if secret == "" {
		return http.StatusBadRequest, errors.New("OnlyOffice JWT secret not configured")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	var claims jwt.MapClaims
	if err = json.Unmarshal(body, &claims); err != nil {
		return http.StatusInternalServerError, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(signed))
	return 0, nil
})

// pathAuthUser extracts a FileBrowser JWT from the first segment of the
// URL path, validates it, and returns the authenticated user along with
// the remaining path.  OnlyOffice strips query parameters from URLs it
// fetches, so the JWT must be embedded in the path instead.
// Expected r.URL.Path (after prefix strip): /<jwt_token>/<file_path>
func pathAuthUser(r *http.Request, d *data) (string, error) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	slashIdx := strings.Index(path, "/")
	if slashIdx < 0 {
		return "", errors.New("invalid URL format")
	}

	tokenStr := path[:slashIdx]
	remaining := path[slashIdx:] // includes leading /

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return d.settings.Key, nil
	}
	var tk authToken
	token, err := jwt.ParseWithClaims(tokenStr, &tk, keyFunc)
	if err != nil || !token.Valid {
		return "", errors.New("unauthorized")
	}

	d.user, err = d.store.Users.Get(d.server.Root, tk.User.ID)
	if err != nil {
		return "", err
	}

	return remaining, nil
}

// rawAuthHandler serves raw file downloads with the auth token in the URL
// path.  OnlyOffice Document Server strips query parameters when fetching
// documents, so the JWT is placed in the path where it is preserved.
// URL format after prefix strip: /<jwt_token>/<file_path>
var rawAuthHandler = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	filePath, err := pathAuthUser(r, d)
	if err != nil {
		return http.StatusUnauthorized, nil
	}

	if !d.user.Perm.Download {
		return http.StatusForbidden, nil
	}

	file, err := files.NewFileInfo(files.FileOptions{
		Fs:         d.user.Fs,
		Path:       filePath,
		Modify:     d.user.Perm.Modify,
		Expand:     false,
		ReadHeader: d.server.TypeDetectionByHeader,
		Checker:    d,
	})
	if err != nil {
		return errToStatus(err), err
	}

	if file.IsDir {
		return http.StatusBadRequest, errors.New("directory download not supported")
	}

	return rawFileHandler(w, r, file)
}

// onlyofficeCallbackAuthHandler handles save callbacks with the auth token
// and file path embedded in the URL path (OnlyOffice strips query params).
// URL format after prefix strip: /<jwt_token>/<save_path>
var onlyofficeCallbackAuthHandler = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	docPath, err := pathAuthUser(r, d)
	if err != nil {
		return http.StatusUnauthorized, nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	var cbData OnlyOfficeCallback
	err = json.Unmarshal(body, &cbData)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if cbData.Status == 2 || cbData.Status == 6 {
		if docPath == "" || docPath == "/" {
			return http.StatusInternalServerError, errors.New("unable to get file save path")
		}

		if !d.user.Perm.Modify || !d.Check(docPath) {
			return http.StatusForbidden, nil
		}

		downloadURL := cbData.URL
		ooExternalURL := d.settings.OnlyOffice.URL
		if ooExternalURL != "" {
			if strings.HasPrefix(ooExternalURL, "/") {
				scheme := "http"
				if r.TLS != nil {
					scheme = "https"
				}
				ooExternalURL = scheme + "://" + r.Host + ooExternalURL
			}
			downloadURL = strings.Replace(downloadURL, ooExternalURL, "http://onlyoffice", 1)
		}

		doc, err := CustomHttpClient.Get(downloadURL)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to download document: %w", err)
		}
		defer doc.Body.Close()

		if doc.StatusCode != http.StatusOK {
			return http.StatusInternalServerError, fmt.Errorf("OnlyOffice returned status %d", doc.StatusCode)
		}

		content, err := io.ReadAll(doc.Body)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to read document body: %w", err)
		}

		if len(content) == 0 {
			return http.StatusInternalServerError, errors.New("OnlyOffice returned empty document, refusing to overwrite")
		}

		err = d.RunHook(func() error {
			_, writeErr := writeFile(d.user.Fs, docPath, bytes.NewReader(content))
			if writeErr != nil {
				return writeErr
			}
			return nil
		}, "save", docPath, "", d.user)

		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	resp := map[string]int{
		"error": 0,
	}
	return renderJSON(w, r, resp)
}
