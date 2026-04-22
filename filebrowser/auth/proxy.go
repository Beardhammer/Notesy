package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// MethodProxyAuth is used to identify no auth.
const MethodProxyAuth settings.AuthMethod = "proxy"

// ProxyAuth is a proxy implementation of an auther.
type ProxyAuth struct {
	Header string `json:"header"`
}

// Auth authenticates the user via an HTTP header. Unknown users are
// auto-created with LockPassword=true so the header is the only way to
// authenticate as them (supports upstream SSO like Authentik).
func (a ProxyAuth) Auth(r *http.Request, usr users.Store, stg *settings.Settings, srv *settings.Server) (*users.User, error) {
	username := r.Header.Get(a.Header)
	if username == "" {
		return nil, os.ErrPermission
	}

	user, err := usr.Get(srv.Root, username)
	if err == nil {
		return user, nil
	}
	if err != errors.ErrNotExist {
		return nil, err
	}

	// First-time login via upstream SSO: create a locked user.
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}
	hashed, err := users.HashPwd(hex.EncodeToString(randomBytes))
	if err != nil {
		return nil, err
	}

	user = &users.User{
		Username:     username,
		Password:     hashed,
		LockPassword: true,
	}
	stg.Defaults.Apply(user)
	// Grant admin so the user can see the assignee list in kanban and
	// otherwise use the full UI. Upstream SSO is the trust boundary, not
	// per-FB-role permission. Tighten via settings if finer control is
	// needed later.
	user.Perm.Admin = true

	home, err := stg.MakeUserDir(user.Username, user.Scope, srv.Root)
	if err != nil {
		return nil, err
	}
	user.Scope = home

	if err := usr.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}

// LoginPage tells that proxy auth doesn't require a login page.
func (a ProxyAuth) LoginPage() bool {
	return false
}
