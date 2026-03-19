package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/zzzzico12/codelens-memory/internal/memory"
)

// ── JSON-RPC types ───────────────────────────────────────

type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ── MCP protocol types ───────────────────────────────────

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type capabilities struct {
	Tools *toolsCap `json:"tools,omitempty"`
}

type toolsCap struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type initializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    capabilities `json:"capabilities"`
	ServerInfo      serverInfo   `json:"serverInfo"`
}

type tool struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema toolSchema `json:"inputSchema"`
}

type toolSchema struct {
	Type       string              `json:"type"`
	Properties map[string]property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

type property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type toolsListResult struct {
	Tools []tool `json:"tools"`
}

type callToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type toolResult struct {
	Content []toolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ── Server ───────────────────────────────────────────────

// Server implements the MCP protocol for CodeLens Memory.
type Server struct {
	engine *memory.Engine
}

// NewServer creates a new MCP server backed by the given memory engine.
func NewServer(engine *memory.Engine) *Server {
	return &Server{engine: engine}
}

// tools returns the list of MCP tools provided by this server.
func (s *Server) tools() []tool {
	return []tool{
		{
			Name:        "memory_search",
			Description: "Search project memories semantically. Use this to find past decisions, conventions, bug fixes, and patterns. Returns the most relevant memories for the given query.",
			InputSchema: toolSchema{
				Type: "object",
				Properties: map[string]property{
					"query": {Type: "string", Description: "Natural language search query (e.g. 'auth architecture decision', 'database indexing', 'naming conventions')"},
					"limit": {Type: "number", Description: "Max results to return (default: 10)"},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "memory_save",
			Description: "Save an important decision, convention, or insight to project memory. Use this when the user makes a significant architectural decision, establishes a coding convention, or solves a tricky bug worth remembering.",
			InputSchema: toolSchema{
				Type: "object",
				Properties: map[string]property{
					"title":    {Type: "string", Description: "Short title summarizing the memory"},
					"content":  {Type: "string", Description: "Detailed description of the decision, convention, or insight"},
					"category": {Type: "string", Description: "Category: decision, convention, bugfix, pattern, or context"},
					"tags":     {Type: "string", Description: "Comma-separated tags for filtering (optional)"},
				},
				Required: []string{"title", "content", "category"},
			},
		},
		{
			Name:        "memory_context",
			Description: "Get relevant project context for the current session. Call this at the start of a session to understand the project's conventions, recent decisions, and important patterns. Returns a curated summary.",
			InputSchema: toolSchema{
				Type: "object",
				Properties: map[string]property{
					"working_dir": {Type: "string", Description: "Current working directory path (optional)"},
					"max_tokens":  {Type: "number", Description: "Maximum tokens for context (default: 2000)"},
				},
			},
		},
		{
			Name:        "memory_stats",
			Description: "Get statistics about the project's memory store. Shows total memories, categories breakdown, database size, and last update time.",
			InputSchema: toolSchema{
				Type: "object",
				Properties: map[string]property{},
			},
		},
	}
}

// handleRequest processes a single JSON-RPC request.
func (s *Server) handleRequest(req jsonrpcRequest) jsonrpcResponse {
	switch req.Method {
	case "initialize":
		return jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: initializeResult{
				ProtocolVersion: "2024-11-05",
				Capabilities: capabilities{
					Tools: &toolsCap{},
				},
				ServerInfo: serverInfo{
					Name:    "codelens-memory",
					Version: "0.1.0",
				},
			},
		}

	case "notifications/initialized":
		// No response needed for notifications
		return jsonrpcResponse{}

	case "tools/list":
		return jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  toolsListResult{Tools: s.tools()},
		}

	case "tools/call":
		var params callToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return s.errorResponse(req.ID, -32602, "invalid params: "+err.Error())
		}
		result := s.callTool(params)
		return jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		}

	default:
		return s.errorResponse(req.ID, -32601, "method not found: "+req.Method)
	}
}

// callTool dispatches to the appropriate tool handler.
func (s *Server) callTool(params callToolParams) toolResult {
	switch params.Name {
	case "memory_search":
		return s.toolSearch(params.Arguments)
	case "memory_save":
		return s.toolSave(params.Arguments)
	case "memory_context":
		return s.toolContext(params.Arguments)
	case "memory_stats":
		return s.toolStats()
	default:
		return toolResult{
			Content: []toolContent{{Type: "text", Text: "Unknown tool: " + params.Name}},
			IsError: true,
		}
	}
}

// ── Tool handlers ────────────────────────────────────────

func (s *Server) toolSearch(args json.RawMessage) toolResult {
	var input struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	json.Unmarshal(args, &input)
	if input.Limit <= 0 {
		input.Limit = 10
	}

	results, err := s.engine.Search(input.Query, input.Limit)
	if err != nil {
		return toolResult{
			Content: []toolContent{{Type: "text", Text: "Search error: " + err.Error()}},
			IsError: true,
		}
	}

	if len(results) == 0 {
		return toolResult{
			Content: []toolContent{{Type: "text", Text: "No memories found for: " + input.Query}},
		}
	}

	text := fmt.Sprintf("Found %d memories:\n\n", len(results))
	for i, r := range results {
		text += fmt.Sprintf("### %d. [%s] %s\n%s\n📅 %s\n\n",
			i+1, r.Category, r.Title, r.Content, r.CreatedAt.Format("2006-01-02"))
	}

	return toolResult{
		Content: []toolContent{{Type: "text", Text: text}},
	}
}

func (s *Server) toolSave(args json.RawMessage) toolResult {
	var input struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Category string `json:"category"`
		Tags     string `json:"tags"`
	}
	json.Unmarshal(args, &input)

	if input.Title == "" || input.Content == "" {
		return toolResult{
			Content: []toolContent{{Type: "text", Text: "title and content are required"}},
			IsError: true,
		}
	}
	if input.Category == "" {
		input.Category = "context"
	}

	id, err := s.engine.Save(input.Title, input.Content, input.Category, "session", "", input.Tags)
	if err != nil {
		return toolResult{
			Content: []toolContent{{Type: "text", Text: "Save error: " + err.Error()}},
			IsError: true,
		}
	}

	return toolResult{
		Content: []toolContent{{Type: "text", Text: fmt.Sprintf("✅ Memory saved (id: %d)\nTitle: %s\nCategory: %s", id, input.Title, input.Category)}},
	}
}

func (s *Server) toolContext(args json.RawMessage) toolResult {
	var input struct {
		WorkingDir string `json:"working_dir"`
		MaxTokens  int    `json:"max_tokens"`
	}
	json.Unmarshal(args, &input)
	if input.MaxTokens <= 0 {
		input.MaxTokens = 2000
	}

	ctx, err := s.engine.Context(input.WorkingDir, input.MaxTokens)
	if err != nil {
		return toolResult{
			Content: []toolContent{{Type: "text", Text: "Context error: " + err.Error()}},
			IsError: true,
		}
	}

	if ctx == "" {
		return toolResult{
			Content: []toolContent{{Type: "text", Text: "No memories yet. Run 'codelens-memory ingest' to learn from Git history, or save memories during coding sessions."}},
		}
	}

	return toolResult{
		Content: []toolContent{{Type: "text", Text: ctx}},
	}
}

func (s *Server) toolStats() toolResult {
	stats, err := s.engine.Stats()
	if err != nil {
		return toolResult{
			Content: []toolContent{{Type: "text", Text: "Stats error: " + err.Error()}},
			IsError: true,
		}
	}

	text := fmt.Sprintf("🧠 CodeLens Memory Stats\n\nTotal memories: %d\n", stats.TotalMemories)
	if len(stats.Categories) > 0 {
		text += "\nCategories:\n"
		for cat, count := range stats.Categories {
			text += fmt.Sprintf("  - %s: %d\n", cat, count)
		}
	}
	text += fmt.Sprintf("\nDatabase size: %s\n", formatBytes(stats.DatabaseSize))
	if !stats.LastUpdated.IsZero() {
		text += fmt.Sprintf("Last updated: %s\n", stats.LastUpdated.Format("2006-01-02 15:04"))
	}

	return toolResult{
		Content: []toolContent{{Type: "text", Text: text}},
	}
}

func (s *Server) errorResponse(id any, code int, msg string) jsonrpcResponse {
	return jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: msg},
	}
}

// ── Stdio transport ──────────────────────────────────────

// ServeStdio runs the MCP server over stdin/stdout (JSON-RPC).
func (s *Server) ServeStdio() error {
	scanner := bufio.NewScanner(os.Stdin)
	// MCP uses newline-delimited JSON
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req jsonrpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			resp := s.errorResponse(nil, -32700, "parse error: "+err.Error())
			s.writeStdout(resp)
			continue
		}

		resp := s.handleRequest(req)
		// Don't respond to notifications (no ID)
		if req.ID == nil && resp.Result == nil && resp.Error == nil {
			continue
		}
		if resp.JSONRPC != "" {
			s.writeStdout(resp)
		}
	}

	return scanner.Err()
}

func (s *Server) writeStdout(resp jsonrpcResponse) {
	data, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stdout, "%s\n", data)
}

// ── SSE transport ────────────────────────────────────────

// ServeSSE runs the MCP server over HTTP with Server-Sent Events.
func (s *Server) ServeSSE(addr string) error {
	mux := http.NewServeMux()

	// SSE endpoint for receiving server events
	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not supported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Send endpoint info
		fmt.Fprintf(w, "event: endpoint\ndata: /message\n\n")
		flusher.Flush()

		// Keep connection alive
		<-r.Context().Done()
	})

	// Message endpoint for receiving requests
	var mu sync.Mutex
	mux.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}

		var req jsonrpcRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		mu.Lock()
		resp := s.handleRequest(req)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	return http.ListenAndServe(addr, mux)
}

func formatBytes(b int64) string {
	switch {
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
