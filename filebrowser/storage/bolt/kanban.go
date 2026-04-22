package bolt

import (
	"github.com/asdine/storm/v3"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/kanban"
)

type kanbanBackend struct {
	db *storm.DB
}

func (s kanbanBackend) All() ([]*kanban.Task, error) {
	var v []*kanban.Task
	err := s.db.All(&v)
	if err == storm.ErrNotFound {
		return v, errors.ErrNotExist
	}

	return v, err
}

func (s kanbanBackend) GetByID(id uint) (*kanban.Task, error) {
	var v kanban.Task
	err := s.db.One("ID", id, &v)
	if err == storm.ErrNotFound {
		return nil, errors.ErrNotExist
	}

	return &v, err
}

func (s kanbanBackend) Save(t *kanban.Task) error {
	return s.db.Save(t)
}

func (s kanbanBackend) Delete(id uint) error {
	err := s.db.DeleteStruct(&kanban.Task{ID: id})
	if err == storm.ErrNotFound {
		return nil
	}
	return err
}

func (s kanbanBackend) AllByBoard(boardID string) ([]*kanban.Task, error) {
	var v []*kanban.Task
	err := s.db.Find("BoardID", boardID, &v)
	if err == storm.ErrNotFound {
		return []*kanban.Task{}, nil
	}
	return v, err
}

func (s kanbanBackend) DeleteByBoard(boardID string) error {
	tasks, err := s.AllByBoard(boardID)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if err := s.db.DeleteStruct(t); err != nil && err != storm.ErrNotFound {
			return err
		}
	}
	return nil
}
