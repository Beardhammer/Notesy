package bolt

import (
	"github.com/asdine/storm/v3"

	"github.com/filebrowser/filebrowser/v2/board"
	"github.com/filebrowser/filebrowser/v2/errors"
)

type boardBackend struct {
	db *storm.DB
}

func (s boardBackend) All() ([]*board.Board, error) {
	var v []*board.Board
	err := s.db.All(&v)
	if err == storm.ErrNotFound {
		return v, errors.ErrNotExist
	}

	return v, err
}

func (s boardBackend) GetByID(id string) (*board.Board, error) {
	var v board.Board
	err := s.db.One("ID", id, &v)
	if err == storm.ErrNotFound {
		return nil, errors.ErrNotExist
	}

	return &v, err
}

func (s boardBackend) Save(b *board.Board) error {
	return s.db.Save(b)
}

func (s boardBackend) Delete(id string) error {
	err := s.db.DeleteStruct(&board.Board{ID: id})
	if err == storm.ErrNotFound {
		return nil
	}
	return err
}
