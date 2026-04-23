// storage/bolt/shout.go
package bolt

import (
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"

	"github.com/filebrowser/filebrowser/v2/shout"
)

type shoutBackend struct {
	db *storm.DB
}

func (s shoutBackend) All() ([]*shout.Message, error) {
	var v []*shout.Message
	err := s.db.AllByIndex("ID", &v)
	if err == storm.ErrNotFound {
		return []*shout.Message{}, nil
	}
	return v, err
}

func (s shoutBackend) Since(id uint) ([]*shout.Message, error) {
	var v []*shout.Message
	err := s.db.Select(q.Gt("ID", id)).OrderBy("ID").Find(&v)
	if err == storm.ErrNotFound {
		return []*shout.Message{}, nil
	}
	return v, err
}

func (s shoutBackend) Save(m *shout.Message) error {
	return s.db.Save(m)
}

func (s shoutBackend) Delete(id uint) error {
	err := s.db.DeleteStruct(&shout.Message{ID: id})
	if err == storm.ErrNotFound {
		return nil
	}
	return err
}

func (s shoutBackend) Count() (int, error) {
	return s.db.Count(&shout.Message{})
}

func (s shoutBackend) OldestIDs(n int) ([]uint, error) {
	if n <= 0 {
		return nil, nil
	}
	var v []*shout.Message
	err := s.db.AllByIndex("ID", &v, storm.Limit(n))
	if err == storm.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ids := make([]uint, len(v))
	for i, m := range v {
		ids[i] = m.ID
	}
	return ids, nil
}
