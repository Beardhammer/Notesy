package bolt

import (
	"github.com/asdine/storm/v3"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/board"
	"github.com/filebrowser/filebrowser/v2/event"
	"github.com/filebrowser/filebrowser/v2/kanban"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/shout"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
)

// NewStorage creates a storage.Storage based on Bolt DB.
func NewStorage(db *storm.DB) (*storage.Storage, error) {
	userStore := users.NewStorage(usersBackend{db: db})
	shareStore := share.NewStorage(shareBackend{db: db})
	settingsStore := settings.NewStorage(settingsBackend{db: db})
	authStore := auth.NewStorage(authBackend{db: db}, userStore)
	kanbanStore := kanban.NewStorage(kanbanBackend{db: db})
	eventStore := event.NewStorage(eventBackend{db: db})
	boardStore := board.NewStorage(boardBackend{db: db})
	shoutStore := shout.NewStorage(shoutBackend{db: db})

	err := save(db, "version", 2) //nolint:gomnd
	if err != nil {
		return nil, err
	}

	return &storage.Storage{
		Auth:     authStore,
		Users:    userStore,
		Share:    shareStore,
		Settings: settingsStore,
		Kanban:   kanbanStore,
		Events:   eventStore,
		Boards:   boardStore,
		Shouts:   shoutStore,
	}, nil
}
