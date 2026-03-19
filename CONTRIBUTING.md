# Contributing to CodeLens Memory

Thank you for your interest in contributing! Here's how to get started.

## Development Setup

```bash
# Clone the repo
git clone https://github.com/zzzzico12/codelens-memory
cd codelens-memory

# Build
go build ./cmd/codelens-memory

# Run tests
go test ./...

# Run with race detector
go test -race ./...
```

## Project Structure

```
codelens-memory/
├── cmd/codelens-memory/    # CLI entrypoint
├── internal/
│   ├── mcp/                # MCP server (JSON-RPC over stdio/SSE)
│   ├── memory/             # Memory engine (search, save, context)
│   ├── ingest/             # Git history parser & classifier
│   ├── embed/              # Embedding providers (Ollama, OpenAI)
│   └── storage/            # SQLite + FTS5 storage layer
├── codelens-memory.toml    # Default configuration
└── go.mod
```

## How to Contribute

### Reporting Bugs

Open an issue with:
- Steps to reproduce
- Expected vs actual behavior
- Your OS and Go version

### Suggesting Features

Open an issue with the `enhancement` label. Describe the use case and why it would be valuable.

### Pull Requests

1. Fork the repo
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Add tests if applicable
5. Run `go test ./...` and ensure everything passes
6. Commit with a descriptive message
7. Push to your fork and open a PR

### Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep functions small and focused
- Add comments for exported types and functions
- Write table-driven tests where possible

## Areas Where Help is Wanted

- 🌐 **Vector search**: Integrate sqlite-vec for semantic embedding search
- 👥 **Team memory**: Git-based shared memory synchronization
- 🖥️ **Web UI**: Browser-based memory viewer
- 📊 **Smarter classification**: Improve commit message categorization with ML
- 🔌 **More embedding providers**: Anthropic, Cohere, local transformers
- 🌍 **i18n**: Japanese and other language support for commit parsing
- 📝 **Documentation**: Tutorials, guides, blog posts

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
