package storage

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Memory represents a single stored memory entry.
type Memory struct {
	ID        int64
	Title     string
	Content   string
	Category  string // decision, convention, bugfix, pattern, context
	Source    string // git-commit, git-pr, session, manual
	SourceRef string // commit hash, PR number, etc.
	Tags      string // comma-separated
	CreatedAt time.Time
	UpdatedAt time.Time
	Score     float64 // populated during search (not stored)
}

// Stats holds aggregate information about the memory store.
type Stats struct {
	TotalMemories int
	Categories    map[string]int
	DatabaseSize  int64
	LastUpdated   time.Time
}

// Store manages the SQLite database for memories.
type Store struct {
	db   *sql.DB
	path string
}

// Open creates or opens a memory database at the given path.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	s := &Store{db: db, path: path}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// migrate creates the schema if it doesn't exist.
func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS memories (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		title      TEXT NOT NULL,
		content    TEXT NOT NULL,
		category   TEXT NOT NULL DEFAULT 'context',
		source     TEXT NOT NULL DEFAULT 'manual',
		source_ref TEXT DEFAULT '',
		tags       TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_memories_category ON memories(category);
	CREATE INDEX IF NOT EXISTS idx_memories_source ON memories(source);
	CREATE INDEX IF NOT EXISTS idx_memories_created_at ON memories(created_at);

	CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
		title,
		content,
		tags,
		content='memories',
		content_rowid='id'
	);

	-- Triggers to keep FTS in sync
	CREATE TRIGGER IF NOT EXISTS memories_ai AFTER INSERT ON memories BEGIN
		INSERT INTO memories_fts(rowid, title, content, tags)
		VALUES (new.id, new.title, new.content, new.tags);
	END;

	CREATE TRIGGER IF NOT EXISTS memories_ad AFTER DELETE ON memories BEGIN
		INSERT INTO memories_fts(memories_fts, rowid, title, content, tags)
		VALUES ('delete', old.id, old.title, old.content, old.tags);
	END;

	CREATE TRIGGER IF NOT EXISTS memories_au AFTER UPDATE ON memories BEGIN
		INSERT INTO memories_fts(memories_fts, rowid, title, content, tags)
		VALUES ('delete', old.id, old.title, old.content, old.tags);
		INSERT INTO memories_fts(rowid, title, content, tags)
		VALUES (new.id, new.title, new.content, new.tags);
	END;
	`
	_, err := s.db.Exec(schema)
	return err
}

// Save inserts a new memory into the store.
func (s *Store) Save(m *Memory) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO memories (title, content, category, source, source_ref, tags, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Title, m.Content, m.Category, m.Source, m.SourceRef, m.Tags,
		m.CreatedAt, time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("insert memory: %w", err)
	}
	return res.LastInsertId()
}

// Search performs full-text search and returns matching memories.
func (s *Store) Search(query string, limit int) ([]Memory, error) {
	rows, err := s.db.Query(`
		SELECT m.id, m.title, m.content, m.category, m.source, m.source_ref,
		       m.tags, m.created_at, m.updated_at,
		       bm25(memories_fts) AS score
		FROM memories_fts fts
		JOIN memories m ON m.id = fts.rowid
		WHERE memories_fts MATCH ?
		ORDER BY score
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer rows.Close()

	var results []Memory
	for rows.Next() {
		var m Memory
		if err := rows.Scan(
			&m.ID, &m.Title, &m.Content, &m.Category, &m.Source,
			&m.SourceRef, &m.Tags, &m.CreatedAt, &m.UpdatedAt, &m.Score,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

// SearchByCategory returns memories filtered by category.
func (s *Store) SearchByCategory(category string, limit int) ([]Memory, error) {
	rows, err := s.db.Query(`
		SELECT id, title, content, category, source, source_ref, tags, created_at, updated_at
		FROM memories
		WHERE category = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, category, limit)
	if err != nil {
		return nil, fmt.Errorf("search by category: %w", err)
	}
	defer rows.Close()

	var results []Memory
	for rows.Next() {
		var m Memory
		if err := rows.Scan(
			&m.ID, &m.Title, &m.Content, &m.Category, &m.Source,
			&m.SourceRef, &m.Tags, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

// Recent returns the most recent memories.
func (s *Store) Recent(limit int) ([]Memory, error) {
	rows, err := s.db.Query(`
		SELECT id, title, content, category, source, source_ref, tags, created_at, updated_at
		FROM memories
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("recent: %w", err)
	}
	defer rows.Close()

	var results []Memory
	for rows.Next() {
		var m Memory
		if err := rows.Scan(
			&m.ID, &m.Title, &m.Content, &m.Category, &m.Source,
			&m.SourceRef, &m.Tags, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

// Stats returns aggregate statistics about the memory store.
func (s *Store) Stats() (*Stats, error) {
	stats := &Stats{
		Categories: make(map[string]int),
	}

	// Total count
	err := s.db.QueryRow("SELECT COUNT(*) FROM memories").Scan(&stats.TotalMemories)
	if err != nil {
		return nil, fmt.Errorf("count: %w", err)
	}

	// Category breakdown
	rows, err := s.db.Query("SELECT category, COUNT(*) FROM memories GROUP BY category")
	if err != nil {
		return nil, fmt.Errorf("categories: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cat string
		var count int
		if err := rows.Scan(&cat, &count); err != nil {
			return nil, err
		}
		stats.Categories[cat] = count
	}

	// Last updated
	var lastUpdated sql.NullTime
	s.db.QueryRow("SELECT MAX(updated_at) FROM memories").Scan(&lastUpdated)
	if lastUpdated.Valid {
		stats.LastUpdated = lastUpdated.Time
	}

	// Database file size
	if info, err := os.Stat(s.path); err == nil {
		stats.DatabaseSize = info.Size()
	}

	return stats, nil
}

// Exists checks if a memory with the given source_ref already exists.
func (s *Store) Exists(sourceRef string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM memories WHERE source_ref = ?", sourceRef).Scan(&count)
	return count > 0, err
}
