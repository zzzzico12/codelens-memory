package memory

import (
	"fmt"
	"time"

	"github.com/zzzzico12/codelens-memory/internal/storage"
)

// Engine is the core memory engine that coordinates search, save, and context.
type Engine struct {
	store *storage.Store
}

// NewEngine creates a new memory engine backed by the given store.
func NewEngine(store *storage.Store) *Engine {
	return &Engine{store: store}
}

// SearchResult wraps a memory with its relevance score.
type SearchResult struct {
	storage.Memory
	Score float64
}

// Search performs a semantic search across all memories.
// Currently uses FTS5 full-text search; will be upgraded to vector search.
func (e *Engine) Search(query string, limit int) ([]SearchResult, error) {
	memories, err := e.store.Search(query, limit)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, len(memories))
	for i, m := range memories {
		results[i] = SearchResult{
			Memory: m,
			Score:  m.Score,
		}
	}
	return results, nil
}

// Save stores a new memory.
func (e *Engine) Save(title, content, category, source, sourceRef, tags string) (int64, error) {
	m := &storage.Memory{
		Title:     title,
		Content:   content,
		Category:  category,
		Source:    source,
		SourceRef: sourceRef,
		Tags:      tags,
		CreatedAt: time.Now(),
	}
	return e.store.Save(m)
}

// SaveIfNew stores a memory only if the source_ref doesn't already exist.
func (e *Engine) SaveIfNew(title, content, category, source, sourceRef, tags string) (int64, bool, error) {
	exists, err := e.store.Exists(sourceRef)
	if err != nil {
		return 0, false, err
	}
	if exists {
		return 0, false, nil
	}
	id, err := e.Save(title, content, category, source, sourceRef, tags)
	return id, true, err
}

// Context returns a curated summary for session start injection.
// It combines recent memories with the most relevant ones for the current project.
func (e *Engine) Context(workingDir string, maxTokens int) (string, error) {
	recent, err := e.store.Recent(5)
	if err != nil {
		return "", fmt.Errorf("recent memories: %w", err)
	}

	decisions, err := e.store.SearchByCategory("decision", 5)
	if err != nil {
		return "", fmt.Errorf("decisions: %w", err)
	}

	conventions, err := e.store.SearchByCategory("convention", 5)
	if err != nil {
		return "", fmt.Errorf("conventions: %w", err)
	}

	// Build context string
	ctx := "# Project Memory (auto-injected by CodeLens Memory)\n\n"

	if len(conventions) > 0 {
		ctx += "## Coding Conventions\n"
		for _, m := range conventions {
			ctx += fmt.Sprintf("- %s: %s\n", m.Title, m.Content)
		}
		ctx += "\n"
	}

	if len(decisions) > 0 {
		ctx += "## Key Decisions\n"
		for _, m := range decisions {
			ctx += fmt.Sprintf("- [%s] %s: %s\n", m.CreatedAt.Format("2006-01-02"), m.Title, m.Content)
		}
		ctx += "\n"
	}

	if len(recent) > 0 {
		ctx += "## Recent Activity\n"
		for _, m := range recent {
			ctx += fmt.Sprintf("- [%s] %s: %s\n", m.CreatedAt.Format("2006-01-02"), m.Title, truncate(m.Content, 200))
		}
		ctx += "\n"
	}

	// Rough token estimation (1 token ≈ 4 chars)
	if len(ctx)/4 > maxTokens {
		ctx = ctx[:maxTokens*4]
	}

	return ctx, nil
}

// Stats returns memory store statistics.
func (e *Engine) Stats() (*storage.Stats, error) {
	return e.store.Stats()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
