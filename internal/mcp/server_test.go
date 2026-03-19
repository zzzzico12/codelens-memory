package mcp

import (
	"encoding/json"
	"testing"

	"github.com/zzzzico12/codelens-memory/internal/memory"
	"github.com/zzzzico12/codelens-memory/internal/storage"
)

func TestHandleInitialize(t *testing.T) {
	server := newTestServer(t)

	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "initialize",
	}

	resp := server.handleRequest(req)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(initializeResult)
	if !ok {
		t.Fatal("result is not initializeResult")
	}
	if result.ServerInfo.Name != "codelens-memory" {
		t.Errorf("expected server name 'codelens-memory', got '%s'", result.ServerInfo.Name)
	}
	if result.ProtocolVersion != "2024-11-05" {
		t.Errorf("expected protocol version '2024-11-05', got '%s'", result.ProtocolVersion)
	}
}

func TestHandleToolsList(t *testing.T) {
	server := newTestServer(t)

	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      float64(2),
		Method:  "tools/list",
	}

	resp := server.handleRequest(req)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(toolsListResult)
	if !ok {
		t.Fatal("result is not toolsListResult")
	}
	if len(result.Tools) != 4 {
		t.Errorf("expected 4 tools, got %d", len(result.Tools))
	}

	names := map[string]bool{}
	for _, tool := range result.Tools {
		names[tool.Name] = true
	}
	for _, expected := range []string{"memory_search", "memory_save", "memory_context", "memory_stats"} {
		if !names[expected] {
			t.Errorf("missing tool: %s", expected)
		}
	}
}

func TestToolSaveAndSearch(t *testing.T) {
	server := newTestServer(t)

	// Save a memory via MCP
	saveArgs, _ := json.Marshal(map[string]string{
		"title":    "Use PostgreSQL",
		"content":  "We chose PostgreSQL over MySQL for better JSON support and full-text search",
		"category": "decision",
		"tags":     "database,postgresql",
	})

	saveReq := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      float64(3),
		Method:  "tools/call",
		Params:  mustMarshal(callToolParams{Name: "memory_save", Arguments: saveArgs}),
	}

	resp := server.handleRequest(saveReq)
	if resp.Error != nil {
		t.Fatalf("save error: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(toolResult)
	if !ok {
		t.Fatal("result is not toolResult")
	}
	if result.IsError {
		t.Fatalf("tool returned error: %s", result.Content[0].Text)
	}

	// Search for the saved memory
	searchArgs, _ := json.Marshal(map[string]any{
		"query": "PostgreSQL database",
		"limit": 5,
	})

	searchReq := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      float64(4),
		Method:  "tools/call",
		Params:  mustMarshal(callToolParams{Name: "memory_search", Arguments: searchArgs}),
	}

	resp = server.handleRequest(searchReq)
	if resp.Error != nil {
		t.Fatalf("search error: %s", resp.Error.Message)
	}

	searchResult, ok := resp.Result.(toolResult)
	if !ok {
		t.Fatal("result is not toolResult")
	}
	if searchResult.IsError {
		t.Fatalf("search tool error: %s", searchResult.Content[0].Text)
	}
	if len(searchResult.Content) == 0 {
		t.Fatal("expected search results")
	}
	t.Logf("search result: %s", searchResult.Content[0].Text)
}

func TestToolStats(t *testing.T) {
	server := newTestServer(t)

	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      float64(5),
		Method:  "tools/call",
		Params:  mustMarshal(callToolParams{Name: "memory_stats", Arguments: json.RawMessage(`{}`)}),
	}

	resp := server.handleRequest(req)
	if resp.Error != nil {
		t.Fatalf("stats error: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(toolResult)
	if !ok {
		t.Fatal("result is not toolResult")
	}
	if result.IsError {
		t.Fatalf("stats tool error: %s", result.Content[0].Text)
	}
	t.Logf("stats: %s", result.Content[0].Text)
}

func TestUnknownMethod(t *testing.T) {
	server := newTestServer(t)

	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      float64(99),
		Method:  "nonexistent/method",
	}

	resp := server.handleRequest(req)
	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected error code -32601, got %d", resp.Error.Code)
	}
}

// ── Helpers ──────────────────────────────────────────────

func newTestServer(t *testing.T) *Server {
	t.Helper()
	path := t.TempDir() + "/test.db"
	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	engine := memory.NewEngine(store)
	return NewServer(engine)
}

func mustMarshal(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
