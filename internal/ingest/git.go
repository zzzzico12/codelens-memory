package ingest

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/zzzzico12/codelens-memory/internal/memory"
)

// GitIngester parses Git history and extracts memories.
type GitIngester struct {
	engine *memory.Engine
}

// NewGitIngester creates a new Git ingester.
func NewGitIngester(engine *memory.Engine) *GitIngester {
	return &GitIngester{engine: engine}
}

// Ingest parses Git history and stores extracted memories.
// Returns the number of new memories created.
func (g *GitIngester) Ingest(since string) (int, error) {
	commits, err := g.getCommits(since)
	if err != nil {
		return 0, fmt.Errorf("get commits: %w", err)
	}

	count := 0
	for _, c := range commits {
		category := classifyCommit(c)
		if category == "" {
			continue // Skip trivial commits
		}

		_, isNew, err := g.engine.SaveIfNew(
			c.subject,
			buildContent(c),
			category,
			"git-commit",
			c.hash,
			buildTags(c),
		)
		if err != nil {
			return count, fmt.Errorf("save commit %s: %w", c.hash[:8], err)
		}
		if isNew {
			count++
		}
	}

	return count, nil
}

// ── Git parsing ──────────────────────────────────────────

type commit struct {
	hash      string
	subject   string
	body      string
	author    string
	date      time.Time
	files     []string
	isMerge   bool
}

func (g *GitIngester) getCommits(since string) ([]commit, error) {
	args := []string{
		"log",
		"--format=%H%n%s%n%b%n%an%n%aI%n---COMMIT_END---",
		"--name-only",
		"--no-merges",
		"-n", "5000",
	}
	if since != "" {
		args = append(args, "--since="+since)
	}

	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	return parseGitLog(string(out)), nil
}

func parseGitLog(output string) []commit {
	blocks := strings.Split(output, "---COMMIT_END---")
	var commits []commit

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		lines := strings.Split(block, "\n")
		if len(lines) < 5 {
			continue
		}

		c := commit{
			hash:    lines[0],
			subject: lines[1],
		}

		// Find the author and date lines (searching from the body area)
		bodyLines := []string{}
		fileLines := []string{}
		foundAuthor := false
		pastFiles := false

		for i := 2; i < len(lines); i++ {
			line := lines[i]
			if !foundAuthor {
				// Try to parse as ISO date to find the date line
				if t, err := time.Parse(time.RFC3339, line); err == nil {
					c.date = t
					if i > 2 {
						c.author = lines[i-1]
					}
					foundAuthor = true
					pastFiles = true
					continue
				}
				bodyLines = append(bodyLines, line)
			} else if pastFiles {
				if strings.TrimSpace(line) == "" {
					continue
				}
				fileLines = append(fileLines, strings.TrimSpace(line))
			}
		}

		if !foundAuthor && len(lines) >= 5 {
			c.author = lines[len(lines)-2]
			t, _ := time.Parse(time.RFC3339, lines[len(lines)-1])
			c.date = t
		}

		c.body = strings.Join(bodyLines, "\n")
		c.files = fileLines
		c.isMerge = strings.HasPrefix(c.subject, "Merge ")

		commits = append(commits, c)
	}

	return commits
}

// ── Classification ───────────────────────────────────────

var (
	decisionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(chose|decided|switched to|migrated to|replaced|instead of)`),
		regexp.MustCompile(`(?i)(architecture|design|approach|strategy)`),
		regexp.MustCompile(`(?i)(because|reason|rationale|trade-?off)`),
	}

	bugfixPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(fix|bug|crash|error|issue|resolve|patch|hotfix)`),
		regexp.MustCompile(`(?i)(root cause|caused by|regression)`),
	}

	conventionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(convention|standard|lint|format|style|naming)`),
		regexp.MustCompile(`(?i)(config|setup|initialize|scaffold)`),
	}

	patternPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(refactor|restructure|reorganize|extract|abstract)`),
		regexp.MustCompile(`(?i)(pattern|middleware|hook|decorator|factory)`),
	}

	// Skip trivial commits
	trivialPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(bump|update|upgrade)\s+(version|deps|dependencies)`),
		regexp.MustCompile(`(?i)^(wip|tmp|temp|test commit)`),
		regexp.MustCompile(`(?i)^merge\s+(branch|pull|remote)`),
	}
)

func classifyCommit(c commit) string {
	text := c.subject + " " + c.body

	// Skip trivial
	for _, p := range trivialPatterns {
		if p.MatchString(c.subject) {
			return ""
		}
	}

	// Score each category
	scores := map[string]int{
		"decision":   matchCount(text, decisionPatterns),
		"bugfix":     matchCount(text, bugfixPatterns),
		"convention": matchCount(text, conventionPatterns),
		"pattern":    matchCount(text, patternPatterns),
	}

	best := ""
	bestScore := 0
	for cat, score := range scores {
		if score > bestScore {
			best = cat
			bestScore = score
		}
	}

	if bestScore == 0 {
		// If the commit message is long enough, it's probably worth keeping
		if len(c.subject+c.body) > 100 {
			return "context"
		}
		return "" // Skip short trivial commits
	}

	// A single weak pattern match on a long descriptive commit is likely
	// incidental — classify as context rather than a specific category.
	if bestScore == 1 && len(c.subject+c.body) > 100 {
		return "context"
	}

	return best
}

func matchCount(text string, patterns []*regexp.Regexp) int {
	count := 0
	for _, p := range patterns {
		if p.MatchString(text) {
			count++
		}
	}
	return count
}

// ── Helpers ──────────────────────────────────────────────

func buildContent(c commit) string {
	content := c.subject
	if c.body != "" {
		content += "\n\n" + strings.TrimSpace(c.body)
	}
	if len(c.files) > 0 {
		content += "\n\nFiles: " + strings.Join(c.files[:min(len(c.files), 10)], ", ")
		if len(c.files) > 10 {
			content += fmt.Sprintf(" (+%d more)", len(c.files)-10)
		}
	}
	return content
}

func buildTags(c commit) string {
	tags := []string{}

	// Extract file extensions as tags
	extMap := map[string]bool{}
	for _, f := range c.files {
		parts := strings.Split(f, ".")
		if len(parts) > 1 {
			ext := parts[len(parts)-1]
			if !extMap[ext] {
				extMap[ext] = true
				tags = append(tags, ext)
			}
		}
	}

	// Extract directory names as tags
	dirMap := map[string]bool{}
	for _, f := range c.files {
		parts := strings.Split(f, "/")
		if len(parts) > 1 && !dirMap[parts[0]] {
			dirMap[parts[0]] = true
			tags = append(tags, parts[0])
		}
	}

	if len(tags) > 10 {
		tags = tags[:10]
	}
	return strings.Join(tags, ",")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
