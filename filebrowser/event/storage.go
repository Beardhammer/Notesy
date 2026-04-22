package event

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

// StorageBackend is the interface to implement for event storage.
type StorageBackend interface {
	All() ([]*Event, error)
	AllByBoard(boardID string) ([]*Event, error)
	GetByID(id uint) (*Event, error)
	Save(e *Event) error
	Delete(id uint) error
	DeleteByBoard(boardID string) error
}

// Storage is an event storage with optional file persistence.
type Storage struct {
	back     StorageBackend
	filePath string
	mu       sync.Mutex
}

// NewStorage creates an event storage from a backend.
func NewStorage(back StorageBackend) *Storage {
	return &Storage{back: back}
}

// SetFilePath sets the path for JSON file persistence.
func (s *Storage) SetFilePath(path string) {
	s.filePath = path
}

// All returns all events.
func (s *Storage) All() ([]*Event, error) {
	return s.back.All()
}

// AllByBoard returns all events for a specific board.
func (s *Storage) AllByBoard(boardID string) ([]*Event, error) {
	return s.back.AllByBoard(boardID)
}

// DeleteByBoard deletes all events for a specific board.
func (s *Storage) DeleteByBoard(boardID string) error {
	if err := s.back.DeleteByBoard(boardID); err != nil {
		return err
	}
	s.persistToFile()
	return nil
}

// GetByID returns an event by its ID.
func (s *Storage) GetByID(id uint) (*Event, error) {
	return s.back.GetByID(id)
}

// Save saves an event and persists to file.
func (s *Storage) Save(e *Event) error {
	if err := s.back.Save(e); err != nil {
		return err
	}
	s.persistToFile()
	return nil
}

// Delete deletes an event by its ID and persists to file.
func (s *Storage) Delete(id uint) error {
	if err := s.back.Delete(id); err != nil {
		return err
	}
	s.persistToFile()
	return nil
}

func (s *Storage) persistToFile() {
	if s.filePath == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	events, err := s.back.All()
	if err != nil {
		log.Printf("[events] failed to read events for persistence: %v", err)
		return
	}
	if events == nil {
		events = []*Event{}
	}

	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		log.Printf("[events] failed to marshal events: %v", err)
		return
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		log.Printf("[events] failed to write %s: %v", s.filePath, err)
	}
}

// LoadFromFile reads events from the JSON file and saves them into the database.
func (s *Storage) LoadFromFile() error {
	if s.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var events []*Event
	if err := json.Unmarshal(data, &events); err != nil {
		return err
	}

	for _, e := range events {
		if err := s.back.Save(e); err != nil {
			log.Printf("[events] failed to load event %d: %v", e.ID, err)
		}
	}

	log.Printf("[events] loaded %d events from %s", len(events), s.filePath)
	return nil
}
