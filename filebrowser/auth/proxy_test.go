package auth

import (
	"net/http/httptest"
	"os"
	"testing"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

type fakeStore struct {
	existing map[string]*users.User
	saved    []*users.User
}

func (f *fakeStore) Get(baseScope string, id interface{}) (*users.User, error) {
	name, ok := id.(string)
	if !ok {
		return nil, fberrors.ErrInvalidDataType
	}
	if u, ok := f.existing[name]; ok {
		return u, nil
	}
	return nil, fberrors.ErrNotExist
}
func (f *fakeStore) Gets(baseScope string) ([]*users.User, error) { return nil, nil }
func (f *fakeStore) Save(u *users.User) error {
	f.saved = append(f.saved, u)
	return nil
}
func (f *fakeStore) Update(u *users.User, fields ...string) error { return nil }
func (f *fakeStore) Delete(id interface{}) error                  { return nil }
func (f *fakeStore) LastUpdate(id uint) int64                     { return 0 }

func TestProxyAuthKnownUser(t *testing.T) {
	store := &fakeStore{existing: map[string]*users.User{"jdoe": {Username: "jdoe"}}}
	stg := &settings.Settings{Defaults: settings.UserDefaults{Scope: "."}, CreateUserDir: true}
	srv := &settings.Server{Root: t.TempDir()}

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Authentik-Username", "jdoe")

	u, err := ProxyAuth{Header: "X-Authentik-Username"}.Auth(r, store, stg, srv)
	if err != nil {
		t.Fatal(err)
	}
	if u.Username != "jdoe" {
		t.Fatal(u)
	}
	if len(store.saved) != 0 {
		t.Fatal("known user should not be saved")
	}
}

func TestProxyAuthEmptyHeaderDenies(t *testing.T) {
	store := &fakeStore{existing: map[string]*users.User{}}
	stg := &settings.Settings{}
	srv := &settings.Server{Root: "/tmp"}
	r := httptest.NewRequest("GET", "/", nil)
	_, err := ProxyAuth{Header: "X-Authentik-Username"}.Auth(r, store, stg, srv)
	if err != os.ErrPermission {
		t.Fatalf("err=%v", err)
	}
}

func TestProxyAuthUnknownUserAutoCreates(t *testing.T) {
	store := &fakeStore{existing: map[string]*users.User{}}
	stg := &settings.Settings{
		Defaults:      settings.UserDefaults{Scope: ".", Perm: users.Permissions{Create: true}},
		CreateUserDir: true,
	}
	srv := &settings.Server{Root: t.TempDir()}
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Authentik-Username", "alice")

	u, err := ProxyAuth{Header: "X-Authentik-Username"}.Auth(r, store, stg, srv)
	if err != nil {
		t.Fatal(err)
	}
	if u.Username != "alice" {
		t.Fatal(u)
	}
	if !u.LockPassword {
		t.Fatal("LockPassword must be true")
	}
	if len(store.saved) != 1 {
		t.Fatalf("saved=%d", len(store.saved))
	}
	if !u.Perm.Create {
		t.Fatal("defaults not applied")
	}
}
