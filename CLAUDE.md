# CLAUDE.md — CodeLens Memory

This file provides context for AI coding agents working on this project.

## Project Overview

CodeLens Memory is a universal memory engine for AI coding agents, implemented as an MCP (Model Context Protocol) server. It gives AI tools persistent memory across sessions by storing and retrieving decisions, conventions, and patterns from a project's history.

## Tech Stack

- **Language**: Go 1.22+
- **Storage**: SQLite with FTS5 full-text search
- **Protocol**: MCP (JSON-RPC over stdio or SSE)
- **Embeddings**: Ollama (local, optional)
- **CLI framework**: cobra

## Architecture

```
cmd/codelens-memory/main.go  → CLI entrypoint (init, serve, ingest, search, stats)
internal/mcp/server.go       → MCP server (JSON-RPC handler, 4 tools)
internal/memory/engine.go    → Core logic (search, save, context injection, stats)
internal/ingest/git.go       → Git history parser & commit classifier
internal/embed/provider.go   → Embedding providers (Ollama, NoOp fallback)
internal/storage/store.go    → SQLite + FTS5 persistence layer
```

## Key Design Decisions

- **MCP-first**: Everything is exposed via MCP tools, not a proprietary plugin API. This ensures compatibility with any MCP client.
- **SQLite + FTS5**: No external database dependencies. Full-text search works without embeddings. Vector search (sqlite-vec) is planned but not yet implemented.
- **Single binary**: No runtime dependencies. `go install` and done.
- **Offline by default**: Ollama for local embeddings. Cloud providers are opt-in.
- **Git-native ingestion**: Learns from commit history automatically via regex-based classification.

## Coding Conventions

- Standard Go formatting (`gofmt`)
- Error wrapping with `fmt.Errorf("context: %w", err)`
- Table-driven tests
- No global state — dependency injection via struct methods
- Internal packages under `internal/` (not importable externally)

## Building & Testing

```bash
go build ./cmd/codelens-memory    # Build
go test -v -race ./...            # Test with race detector
make build                        # Build with version info
```

## Roadmap / TODOs

- [ ] sqlite-vec integration for semantic vector search
- [ ] Config file loading (codelens-memory.toml)
- [ ] Team memory via Git remote sync
- [ ] Web UI for browsing memories
- [ ] Sensitive data redaction (API keys, tokens)
- [ ] Memory decay / relevance scoring by age
