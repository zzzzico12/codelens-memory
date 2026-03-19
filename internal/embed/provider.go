package embed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Provider generates vector embeddings for text.
type Provider interface {
	Embed(text string) ([]float32, error)
	EmbedBatch(texts []string) ([][]float32, error)
}

// OllamaProvider generates embeddings using a local Ollama instance.
type OllamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaProvider creates a provider that talks to a local Ollama server.
func NewOllamaProvider(baseURL, model string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "nomic-embed-text"
	}
	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

type ollamaEmbedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type ollamaEmbedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// Embed generates an embedding for a single text.
func (o *OllamaProvider) Embed(text string) ([]float32, error) {
	body, err := json.Marshal(ollamaEmbedRequest{
		Model: o.model,
		Input: text,
	})
	if err != nil {
		return nil, err
	}

	resp, err := o.client.Post(o.baseURL+"/api/embed", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama request: %w (is Ollama running?)", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(respBody))
	}

	var result ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return result.Embeddings[0], nil
}

// EmbedBatch generates embeddings for multiple texts.
func (o *OllamaProvider) EmbedBatch(texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := o.Embed(text)
		if err != nil {
			return nil, fmt.Errorf("embed text %d: %w", i, err)
		}
		results[i] = emb
	}
	return results, nil
}

// ── NoOp Provider (for when no embedding service is available) ──

// NoOpProvider is a fallback that returns empty embeddings.
// The system will fall back to FTS5 full-text search when this is used.
type NoOpProvider struct{}

// NewNoOpProvider creates a no-op embedding provider.
func NewNoOpProvider() *NoOpProvider {
	return &NoOpProvider{}
}

// Embed returns nil (FTS5 will be used instead).
func (n *NoOpProvider) Embed(text string) ([]float32, error) {
	return nil, nil
}

// EmbedBatch returns nil for all texts.
func (n *NoOpProvider) EmbedBatch(texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	return results, nil
}
