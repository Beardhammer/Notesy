package board

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

// StorageBackend is the interface to implement for board storage.
type StorageBackend interface {
	All() ([]*Board, error)
	GetByID(id string) (*Board, error)
	Save(b *Board) error
	Delete(id string) error
}

// Storage is a board storage with optional file persistence.
type Storage struct {
	back     StorageBackend
	filePath string
	mu       sync.Mutex
}

// NewStorage creates a board storage from a backend.
func NewStorage(back StorageBackend) *Storage {
	return &Storage{back: back}
}

// SetFilePath sets the path for JSON file persistence.
func (s *Storage) SetFilePath(path string) {
	s.filePath = path
}

// All returns all boards.
func (s *Storage) All() ([]*Board, error) {
	return s.back.All()
}

// GetByID returns a board by its ID.
func (s *Storage) GetByID(id string) (*Board, error) {
	return s.back.GetByID(id)
}

// Save saves a board and persists to file.
func (s *Storage) Save(b *Board) error {
	if err := s.back.Save(b); err != nil {
		return err
	}
	s.persistToFile()
	return nil
}

// Delete deletes a board by its ID and persists to file.
func (s *Storage) Delete(id string) error {
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

	boards, err := s.back.All()
	if err != nil {
		log.Printf("[boards] failed to read boards for persistence: %v", err)
		return
	}
	if boards == nil {
		boards = []*Board{}
	}

	data, err := json.MarshalIndent(boards, "", "  ")
	if err != nil {
		log.Printf("[boards] failed to marshal boards: %v", err)
		return
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		log.Printf("[boards] failed to write %s: %v", s.filePath, err)
	}
}

// LoadFromFile reads boards from the JSON file and saves them into the database.
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

	var boards []*Board
	if err := json.Unmarshal(data, &boards); err != nil {
		return err
	}

	for _, b := range boards {
		if err := s.back.Save(b); err != nil {
			log.Printf("[boards] failed to load board %s: %v", b.ID, err)
		}
	}

	log.Printf("[boards] loaded %d boards from %s", len(boards), s.filePath)
	return nil
}
