package memory

import (
	"testing"

	"github.com/zzzzico12/codelens-memory/internal/storage"
)

func TestEngineSaveAndSearch(t *testing.T) {
	engine := newTestEngine(t)

	id, err := engine.Save(
		"Use kebab-case for files",
		"All component files should use kebab-case naming: my-component.tsx",
		"convention", "session", "", "naming,files",
	)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	results, err := engine.Search("file naming convention", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if results[0].Title != "Use kebab-case for files" {
		t.Errorf("wrong title: %s", results[0].Title)
	}
}

func TestEngineSaveIfNew(t *testing.T) {
	engine := newTestEngine(t)

	// First save should succeed
	_, isNew, err := engine.SaveIfNew("test", "content", "context", "git-commit", "abc123", "")
	if err != nil {
		t.Fatalf("SaveIfNew: %v", err)
	}
	if !isNew {
		t.Fatal("expected isNew=true for first save")
	}

	// Second save with same source_ref should be skipped
	_, isNew, err = engine.SaveIfNew("test", "content", "context", "git-commit", "abc123", "")
	if err != nil {
		t.Fatalf("SaveIfNew: %v", err)
	}
	if isNew {
		t.Fatal("expected isNew=false for duplicate")
	}
}

func TestEngineContext(t *testing.T) {
	engine := newTestEngine(t)

	// Save some memories in different categories
	if _, err := engine.Save("Use 2-space indent", "All files use 2-space indentation", "convention", "session", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Save("Choose Redis for cache", "Redis chosen over Memcached for pub/sub support", "decision", "session", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Save("Recent work on auth", "Implemented OAuth2 PKCE flow for mobile app", "context", "session", "", ""); err != nil {
		t.Fatal(err)
	}

	ctx, err := engine.Context(".", 2000)
	if err != nil {
		t.Fatalf("Context: %v", err)
	}
	if ctx == "" {
		t.Fatal("expected non-empty context")
	}

	// Check that context includes sections
	if !containsStr(ctx, "Coding Conventions") {
		t.Error("context should include Coding Conventions section")
	}
	if !containsStr(ctx, "Key Decisions") {
		t.Error("context should include Key Decisions section")
	}
	t.Logf("context:\n%s", ctx)
}

func TestEngineStats(t *testing.T) {
	engine := newTestEngine(t)

	if _, err := engine.Save("d1", "content", "decision", "session", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Save("b1", "content", "bugfix", "session", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Save("b2", "content", "bugfix", "session", "", ""); err != nil {
		t.Fatal(err)
	}

	stats, err := engine.Stats()
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.TotalMemories != 3 {
		t.Errorf("expected 3 memories, got %d", stats.TotalMemories)
	}
	if stats.Categories["bugfix"] != 2 {
		t.Errorf("expected 2 bugfixes, got %d", stats.Categories["bugfix"])
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is a ..."},
		{"exact", 5, "exact"},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.expected)
		}
	}
}

// ── Helpers ──────────────────────────────────────────────

func newTestEngine(t *testing.T) *Engine {
	t.Helper()
	path := t.TempDir() + "/test.db"
	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return NewEngine(store)
}

func containsStr(haystack, needle string) bool {
	return len(haystack) >= len(needle) && searchStr(haystack, needle)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
