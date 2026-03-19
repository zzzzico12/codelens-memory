<p align="center">
  <img src="assets/logo.svg" width="120" alt="CodeLens Memory" />
</p>

<h1 align="center">CodeLens Memory</h1>

<p align="center">
  <strong>Your AI coding tools forget everything. This fixes that.</strong>
</p>

<p align="center">
  A universal memory engine for AI coding agents.<br />
  Works with <b>any</b> MCP-compatible tool — Claude Code, Cursor, Windsurf, Cline, and more.
</p>

<p align="center">
  <a href="#quick-start">Quick Start</a> •
  <a href="#why-codelens-memory">Why?</a> •
  <a href="#how-it-works">How It Works</a> •
  <a href="#features">Features</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#contributing">Contributing</a>
</p>

<p align="center">
  <a href="https://github.com/yourname/codelens-memory/releases"><img src="https://img.shields.io/github/v/release/yourname/codelens-memory?style=flat-square&color=7F77DD" alt="Release" /></a>
  <a href="https://github.com/yourname/codelens-memory/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue?style=flat-square" alt="License" /></a>
  <a href="https://github.com/yourname/codelens-memory/stargazers"><img src="https://img.shields.io/github/stars/yourname/codelens-memory?style=flat-square&color=f5c542" alt="Stars" /></a>
</p>

---

## The problem

Every AI coding tool — Claude Code, Cursor, Windsurf — starts every session with **zero memory** of your project.

> **You:** "What did we decide about the auth architecture?"
>
> **AI:** "I don't have context from previous sessions."

You've explained your coding conventions 50 times. You've re-described your database schema in every session. You've watched your AI make the same mistake it made yesterday, because it doesn't remember yesterday.

**CodeLens Memory gives your AI a persistent brain that works across every tool.**

## Why CodeLens Memory

Other memory tools exist, but they're all locked to a single tool:

| | CodeLens Memory | claude-mem | claude-brain | Cursor Memories |
|---|:---:|:---:|:---:|:---:|
| **Claude Code** | ✅ | ✅ | ✅ | ❌ |
| **Cursor** | ✅ | ❌ | ❌ | ✅ |
| **Windsurf / Cline / OpenCode** | ✅ | ❌ | ❌ | ❌ |
| **Any future MCP client** | ✅ | ❌ | ❌ | ❌ |
| **Self-hosted / offline** | ✅ | ✅ | ✅ | ❌ |
| **Auto-learns from Git** | ✅ | ❌ | ❌ | ❌ |
| **Zero dependencies** | ✅ | ❌ | ✅ | — |
| **Portable single file** | ✅ | ❌ | ✅ | ❌ |
| **Cross-project memory** | ✅ | ❌ | ❌ | partial |

**CodeLens Memory is the only tool that works everywhere, because it speaks MCP — the universal protocol adopted by OpenAI, Google, Microsoft, and the Linux Foundation.**

## Quick Start

### Install

```bash
# macOS / Linux
curl -fsSL https://codelens-memory.dev/install.sh | bash

# Or with Go
go install github.com/yourname/codelens-memory@latest
```

### Connect to Claude Code

```bash
# Add as MCP server (one-time setup)
claude mcp add codelens-memory -- codelens-memory serve
```

### Connect to Cursor

Add to your `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "codelens-memory": {
      "command": "codelens-memory",
      "args": ["serve"]
    }
  }
}
```

### Initialize in your project

```bash
cd your-project
codelens-memory init    # Creates .codelens/memory.db
codelens-memory ingest  # Learns from Git history
```

**That's it.** Your AI now remembers everything.

## How It Works

```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ Claude Code  │  │   Cursor    │  │  Windsurf   │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │
       └────────────┬────┴────────┬────────┘
                    │  MCP Protocol │
                    ▼              ▼
          ┌─────────────────────────────┐
          │    CodeLens Memory Server    │
          │                             │
          │  memory_search   "auth"  →  │──▶ Semantic search
          │  memory_save     {...}   →  │──▶ Store decision
          │  memory_context  auto    →  │──▶ Inject at start
          │  memory_stats    ...     →  │──▶ Overview
          └──────────────┬──────────────┘
                         │
          ┌──────────────┴──────────────┐
          │                             │
          ▼                             ▼
   ┌─────────────┐            ┌──────────────┐
   │  Git History │            │  memory.db   │
   │  commits     │            │  SQLite +    │
   │  diffs       │            │  sqlite-vec  │
   │  PR messages │            │  (single     │
   └─────────────┘            │   file)      │
                               └──────────────┘
```

### The Memory Pipeline

1. **Ingest** — Parses your Git history (commits, diffs, PR messages) and extracts decisions, patterns, and conventions.
2. **Observe** — During coding sessions, captures important decisions and architectural choices in real time.
3. **Recall** — When your AI needs context, `memory_search` retrieves the most relevant memories using semantic search.
4. **Inject** — At session start, `memory_context` automatically provides relevant background so the AI already knows your project.

### What Gets Remembered

- 🏗️ **Architectural decisions** — "We chose JWT over sessions because..."
- 🐛 **Bug fixes & root causes** — "This crash was caused by..."
- 📏 **Coding conventions** — "We use kebab-case for file names and..."
- 🔧 **Configuration choices** — "PostgreSQL over MySQL because..."
- 📝 **PR discussions** — Key decisions extracted from review threads
- ⚡ **Session insights** — What you told the AI during coding sessions

## Features

### 🔌 Universal MCP Server

Speaks the Model Context Protocol, so it works with any compatible tool — today and in the future. No vendor lock-in.

### 🧠 Git-Native Intelligence

Doesn't just record what you tell it — it **learns from your Git history**. Run `codelens-memory ingest` and it extracts years of decisions from your commits and PRs.

### 📦 Single Binary, Zero Dependencies

One `go install` and you're done. No Python, no Node.js, no Docker, no ChromaDB, no external databases. The memory lives in a single `.db` file.

### 🔒 Fully Offline

Embeddings generated locally via Ollama (or any OpenAI-compatible API). Your code and decisions never leave your machine.

### 💾 Portable Memory

Your `.codelens/memory.db` is a single file. Copy it, `git commit` it, `scp` it to another machine. Your AI's memory travels with your project.

### 🌐 Cross-Project Memory

Optionally share memories across projects. That auth pattern you figured out in Project A? Your AI remembers it in Project B.

### 👥 Team Memory (coming soon)

Share a memory store across your team. When a teammate solves a tricky deployment issue, everyone's AI knows about it.

## MCP Tools

CodeLens Memory exposes 4 tools via MCP:

### `memory_search`

Semantic search across all memories.

```
"What did we decide about database indexing?"
→ Returns relevant decisions, code patterns, and their context
```

### `memory_save`

Explicitly save an important decision or insight.

```
"Remember: we chose Redis for session storage because
PostgreSQL's advisory locks caused deadlocks under load."
```

### `memory_context`

Auto-called at session start. Returns a curated summary of the most relevant context for the current working directory and recent files.

### `memory_stats`

Overview of your memory store: total memories, topics, last updated, storage size.

## Configuration

### `codelens-memory.toml`

```toml
[memory]
# Where to store the memory database
path = ".codelens/memory.db"

# Cross-project shared memory (optional)
# shared_path = "~/.codelens/shared.db"

[embeddings]
# "ollama" (default, local) | "openai" | "anthropic"
provider = "ollama"

# Model for generating embeddings
model = "nomic-embed-text"

# Only needed for cloud providers
# api_key = "sk-..."

[ingest]
# Max commits to process on first ingest
max_commits = 5000

# Include PR/merge commit messages
include_merge_commits = true

# File patterns to ignore
ignore_patterns = ["*.lock", "*.min.js", "node_modules/**"]

[context]
# Max tokens to inject at session start
max_context_tokens = 2000

# How many memories to consider
top_k = 10
```

## CLI Reference

```bash
codelens-memory init                    # Initialize in current project
codelens-memory serve                   # Start MCP server (stdio mode)
codelens-memory serve --sse :8377       # Start MCP server (SSE mode)
codelens-memory ingest                  # Learn from Git history
codelens-memory ingest --since 6months  # Learn from recent history only
codelens-memory search "auth"           # Search memories from terminal
codelens-memory stats                   # Show memory statistics
codelens-memory export > memories.json  # Export all memories
codelens-memory import < memories.json  # Import memories
codelens-memory prune --older 1year     # Clean old memories
```

## Privacy & Security

- **100% local by default.** No data is sent anywhere unless you configure a cloud embedding provider.
- **You own your data.** The memory file is plain SQLite — inspect it, query it, delete it anytime.
- **Gitignore-friendly.** `codelens-memory init` adds `.codelens/` to your `.gitignore`. Commit the memory only if you choose to.
- **Sensitive data filtering.** Automatically redacts API keys, tokens, and passwords from memories.

## Roadmap

- [x] Core MCP server with 4 tools
- [x] Git history ingestion
- [x] SQLite + sqlite-vec storage
- [x] Ollama embedding support
- [ ] Team memory sharing via Git remote
- [ ] Web UI for browsing memories
- [ ] VS Code extension for memory visualization
- [ ] Automatic convention detection from code patterns
- [ ] Memory decay (older memories gradually lose relevance)
- [ ] Plugin system for custom memory sources (Jira, Slack, Notion)

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

```bash
git clone https://github.com/yourname/codelens-memory
cd codelens-memory
go build ./cmd/codelens-memory
go test ./...
```

### Project Structure

```
codelens-memory/
├── cmd/
│   └── codelens-memory/    # CLI entrypoint
├── internal/
│   ├── mcp/                # MCP server implementation
│   ├── memory/             # Memory engine (search, save, context)
│   ├── ingest/             # Git history parser
│   ├── embed/              # Embedding providers (ollama, openai)
│   └── storage/            # SQLite + sqlite-vec layer
├── codelens-memory.toml    # Default config
└── go.mod
```

## License

MIT — use it however you want.

---

<p align="center">
  <strong>Stop re-explaining your project to your AI. Give it a memory.</strong>
</p>

<p align="center">
  <a href="#quick-start">Get Started →</a>
</p>
