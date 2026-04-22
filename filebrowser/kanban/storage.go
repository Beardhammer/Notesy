package kanban

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

// StorageBackend is the interface to implement for kanban task storage.
type StorageBackend interface {
	All() ([]*Task, error)
	AllByBoard(boardID string) ([]*Task, error)
	GetByID(id uint) (*Task, error)
	Save(t *Task) error
	Delete(id uint) error
	DeleteByBoard(boardID string) error
}

// Storage is a kanban task storage with optional file persistence.
type Storage struct {
	back     StorageBackend
	filePath string
	mu       sync.Mutex
}

// NewStorage creates a kanban storage from a backend.
func NewStorage(back StorageBackend) *Storage {
	return &Storage{back: back}
}

// SetFilePath sets the path for JSON file persistence.
func (s *Storage) SetFilePath(path string) {
	s.filePath = path
}

// All returns all kanban tasks.
func (s *Storage) All() ([]*Task, error) {
	return s.back.All()
}

// AllByBoard returns all tasks for a specific board.
func (s *Storage) AllByBoard(boardID string) ([]*Task, error) {
	return s.back.AllByBoard(boardID)
}

// DeleteByBoard deletes all tasks for a specific board.
func (s *Storage) DeleteByBoard(boardID string) error {
	if err := s.back.DeleteByBoard(boardID); err != nil {
		return err
	}
	s.persistToFile()
	return nil
}

// GetByID returns a kanban task by its ID.
func (s *Storage) GetByID(id uint) (*Task, error) {
	return s.back.GetByID(id)
}

// Save saves a kanban task and persists to file.
func (s *Storage) Save(t *Task) error {
	if err := s.back.Save(t); err != nil {
		return err
	}
	s.persistToFile()
	return nil
}

// Delete deletes a kanban task by its ID and persists to file.
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

	tasks, err := s.back.All()
	if err != nil {
		log.Printf("[kanban] failed to read tasks for persistence: %v", err)
		return
	}
	if tasks == nil {
		tasks = []*Task{}
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		log.Printf("[kanban] failed to marshal tasks: %v", err)
		return
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		log.Printf("[kanban] failed to write %s: %v", s.filePath, err)
	}
}

// LoadFromFile reads tasks from the JSON file and saves them into the database.
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

	var tasks []*Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return err
	}

	for _, t := range tasks {
		if err := s.back.Save(t); err != nil {
			log.Printf("[kanban] failed to load task %d: %v", t.ID, err)
		}
	}

	log.Printf("[kanban] loaded %d tasks from %s", len(tasks), s.filePath)
	return nil
}
