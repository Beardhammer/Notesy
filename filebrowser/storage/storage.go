package storage

import (
	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/board"
	"github.com/filebrowser/filebrowser/v2/event"
	"github.com/filebrowser/filebrowser/v2/kanban"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/shout"
	"github.com/filebrowser/filebrowser/v2/users"
)

// Storage is a storage powered by a Backend which makes the necessary
// verifications when fetching and saving data to ensure consistency.
type Storage struct {
	Users    users.Store
	Share    *share.Storage
	Auth     *auth.Storage
	Settings *settings.Storage
	Kanban   *kanban.Storage
	Events   *event.Storage
	Boards   *board.Storage
	Shouts   *shout.Storage
}
