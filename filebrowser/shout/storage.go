// shout/storage.go
package shout

import "sync"

// StorageBackend is the interface to implement for shout storage.
type StorageBackend interface {
	All() ([]*Message, error)
	Since(id uint) ([]*Message, error) // messages with ID > id, ordered ascending
	Save(m *Message) error
	Delete(id uint) error
	Count() (int, error)
	OldestIDs(n int) ([]uint, error) // n oldest IDs ascending; used by trim
}

// Storage wraps a backend and (later) a hub.
type Storage struct {
	back StorageBackend
	hub  *Hub
	mu   sync.Mutex
}

// NewStorage creates a shout storage.
func NewStorage(back StorageBackend) *Storage {
	return &Storage{back: back, hub: NewHub()}
}

func (s *Storage) All() ([]*Message, error)          { return s.back.All() }
func (s *Storage) Since(id uint) ([]*Message, error) { return s.back.Since(id) }
func (s *Storage) Hub() *Hub                         { return s.hub }

// MaxMessages is the rolling cap on stored shouts.
const MaxMessages = 200

func (s *Storage) Save(m *Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.back.Save(m); err != nil {
		return err
	}

	count, err := s.back.Count()
	if err != nil {
		return err
	}
	if count <= MaxMessages {
		return nil
	}

	ids, err := s.back.OldestIDs(count - MaxMessages)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := s.back.Delete(id); err != nil {
			return err
		}
	}
	return nil
}
