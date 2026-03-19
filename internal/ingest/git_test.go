package ingest

import (
	"testing"
	"time"
)

func TestClassifyCommit(t *testing.T) {
	tests := []struct {
		name     string
		commit   commit
		expected string
	}{
		{
			name:     "decision commit",
			commit:   commit{subject: "Switch from sessions to JWT because API is stateless", body: "We decided to use JWT tokens instead of server-side sessions for the REST API."},
			expected: "decision",
		},
		{
			name:     "bugfix commit",
			commit:   commit{subject: "Fix crash when user has no profile", body: "Root cause: nil pointer dereference in user.Profile()"},
			expected: "bugfix",
		},
		{
			name:     "convention commit",
			commit:   commit{subject: "Add ESLint config for consistent formatting", body: "Standardize on single quotes and 2-space indentation"},
			expected: "convention",
		},
		{
			name:     "pattern commit",
			commit:   commit{subject: "Refactor auth into middleware pattern", body: "Extract authentication logic into a reusable middleware"},
			expected: "pattern",
		},
		{
			name:     "trivial version bump",
			commit:   commit{subject: "Bump version to 1.2.3"},
			expected: "",
		},
		{
			name:     "trivial wip",
			commit:   commit{subject: "wip"},
			expected: "",
		},
		{
			name:     "long context commit",
			commit:   commit{subject: "Add comprehensive user onboarding flow with email verification, profile setup, and team invitation. This includes the new welcome screen component and tutorial overlay."},
			expected: "context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyCommit(tt.commit)
			if got != tt.expected {
				t.Errorf("classifyCommit() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseGitLog(t *testing.T) {
	output := `abc1234567890
Fix auth token expiry bug
The token was not being refreshed when it expired during an active session.
Root cause was missing check in middleware.
John Doe
2026-01-15T10:30:00+09:00

internal/auth/middleware.go
internal/auth/token.go
---COMMIT_END---
def9876543210
Add user profile page
Basic profile page with avatar upload
Jane Smith
2026-01-14T09:00:00+09:00

web/pages/profile.tsx
web/components/avatar.tsx
---COMMIT_END---`

	commits := parseGitLog(output)
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}

	if commits[0].hash != "abc1234567890" {
		t.Errorf("wrong hash: %s", commits[0].hash)
	}
	if commits[0].subject != "Fix auth token expiry bug" {
		t.Errorf("wrong subject: %s", commits[0].subject)
	}
	if commits[0].author != "John Doe" {
		t.Errorf("wrong author: %s", commits[0].author)
	}
	if commits[0].date.Before(time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("wrong date: %v", commits[0].date)
	}
}

func TestBuildTags(t *testing.T) {
	c := commit{
		files: []string{
			"internal/auth/middleware.go",
			"internal/auth/token.go",
			"web/pages/profile.tsx",
		},
	}

	tags := buildTags(c)
	if tags == "" {
		t.Fatal("expected non-empty tags")
	}

	// Should contain file extensions and directory names
	// The exact format depends on implementation but should be comma-separated
	t.Logf("tags: %s", tags)
}
