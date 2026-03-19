package storage

import (
	"os"
	"testing"
	"time"
)

func TestOpenAndMigrate(t *testing.T) {
	path := t.TempDir() + "/test.db"
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer store.Close()

	// Verify file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("database file not created")
	}
}

func TestSaveAndSearch(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	// Save a memory
	m := &Memory{
		Title:     "JWT auth decision",
		Content:   "We chose JWT over sessions because our API is stateless and serves mobile clients",
		Category:  "decision",
		Source:    "session",
		SourceRef: "test-1",
		Tags:      "auth,jwt,api",
		CreatedAt: time.Now(),
	}

	id, err := store.Save(m)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	// Search for it
	results, err := store.Search("JWT auth", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if results[0].Title != "JWT auth decision" {
		t.Errorf("expected title 'JWT auth decision', got '%s'", results[0].Title)
	}
}

func TestSearchByCategory(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	// Save memories in different categories
	if _, err := store.Save(&Memory{Title: "Decision 1", Content: "content", Category: "decision", Source: "test", CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(&Memory{Title: "Bugfix 1", Content: "content", Category: "bugfix", Source: "test", CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(&Memory{Title: "Decision 2", Content: "content", Category: "decision", Source: "test", CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	results, err := store.SearchByCategory("decision", 10)
	if err != nil {
		t.Fatalf("SearchByCategory: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 decisions, got %d", len(results))
	}
}

func TestExists(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	if _, err := store.Save(&Memory{Title: "test", Content: "content", Category: "context", Source: "git-commit", SourceRef: "abc123", CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	exists, err := store.Exists("abc123")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !exists {
		t.Fatal("expected memory to exist")
	}

	exists, err = store.Exists("nonexistent")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if exists {
		t.Fatal("expected memory to not exist")
	}
}

func TestStats(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	if _, err := store.Save(&Memory{Title: "d1", Content: "c", Category: "decision", Source: "test", CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(&Memory{Title: "b1", Content: "c", Category: "bugfix", Source: "test", CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	stats, err := store.Stats()
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.TotalMemories != 2 {
		t.Errorf("expected 2 total, got %d", stats.TotalMemories)
	}
	if stats.Categories["decision"] != 1 {
		t.Errorf("expected 1 decision, got %d", stats.Categories["decision"])
	}
	if stats.Categories["bugfix"] != 1 {
		t.Errorf("expected 1 bugfix, got %d", stats.Categories["bugfix"])
	}
}

func openTestStore(t *testing.T) *Store {
	t.Helper()
	path := t.TempDir() + "/test.db"
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return store
}
