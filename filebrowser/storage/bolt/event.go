package bolt

import (
	"github.com/asdine/storm/v3"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/event"
)

type eventBackend struct {
	db *storm.DB
}

func (s eventBackend) All() ([]*event.Event, error) {
	var v []*event.Event
	err := s.db.All(&v)
	if err == storm.ErrNotFound {
		return v, errors.ErrNotExist
	}

	return v, err
}

func (s eventBackend) GetByID(id uint) (*event.Event, error) {
	var v event.Event
	err := s.db.One("ID", id, &v)
	if err == storm.ErrNotFound {
		return nil, errors.ErrNotExist
	}

	return &v, err
}

func (s eventBackend) Save(e *event.Event) error {
	return s.db.Save(e)
}

func (s eventBackend) Delete(id uint) error {
	err := s.db.DeleteStruct(&event.Event{ID: id})
	if err == storm.ErrNotFound {
		return nil
	}
	return err
}

func (s eventBackend) AllByBoard(boardID string) ([]*event.Event, error) {
	var v []*event.Event
	err := s.db.Find("BoardID", boardID, &v)
	if err == storm.ErrNotFound {
		return []*event.Event{}, nil
	}
	return v, err
}

func (s eventBackend) DeleteByBoard(boardID string) error {
	events, err := s.AllByBoard(boardID)
	if err != nil {
		return err
	}
	for _, e := range events {
		if err := s.db.DeleteStruct(e); err != nil && err != storm.ErrNotFound {
			return err
		}
	}
	return nil
}
