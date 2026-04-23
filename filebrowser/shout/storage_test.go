package shout_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/asdine/storm/v3"

	"github.com/filebrowser/filebrowser/v2/shout"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
)

func newTestStore(t *testing.T) *shout.Storage {
	t.Helper()
	dir := t.TempDir()
	db, err := storm.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close(); os.RemoveAll(dir) })
	store, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	return store.Shouts
}

func TestSaveStoresMessage(t *testing.T) {
	s := newTestStore(t)
	m := &shout.Message{Author: "alice", Body: "hi", CreatedAt: 1}
	if err := s.Save(m); err != nil {
		t.Fatalf("save: %v", err)
	}
	if m.ID == 0 {
		t.Fatalf("expected auto-increment ID, got 0")
	}
}

func TestSaveTrimsTo200(t *testing.T) {
	s := newTestStore(t)
	for i := 0; i < 205; i++ {
		if err := s.Save(&shout.Message{Author: "a", Body: fmt.Sprintf("m%d", i), CreatedAt: int64(i)}); err != nil {
			t.Fatalf("save %d: %v", i, err)
		}
	}
	all, err := s.All()
	if err != nil {
		t.Fatalf("all: %v", err)
	}
	if len(all) != 200 {
		t.Fatalf("expected 200 messages, got %d", len(all))
	}
	// Oldest remaining should be m5 (ID 6) because IDs 1..5 were trimmed.
	if all[0].Body != "m5" {
		t.Fatalf("expected oldest body m5, got %q", all[0].Body)
	}
}

func TestSaveUnderCapDoesNotTrim(t *testing.T) {
	s := newTestStore(t)
	for i := 0; i < 5; i++ {
		_ = s.Save(&shout.Message{Author: "a", Body: fmt.Sprintf("m%d", i)})
	}
	all, _ := s.All()
	if len(all) != 5 {
		t.Fatalf("expected 5 messages, got %d", len(all))
	}
}

func TestSinceReturnsOnlyNewer(t *testing.T) {
	s := newTestStore(t)
	for i := 0; i < 3; i++ {
		_ = s.Save(&shout.Message{Author: "a", Body: fmt.Sprintf("m%d", i)})
	}
	out, err := s.Since(1)
	if err != nil {
		t.Fatalf("since: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(out))
	}
	if out[0].ID != 2 || out[1].ID != 3 {
		t.Fatalf("expected IDs 2,3, got %d,%d", out[0].ID, out[1].ID)
	}
}
