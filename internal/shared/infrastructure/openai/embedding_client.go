package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EmbeddingClient handles OpenAI embedding API calls
type EmbeddingClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

// EmbeddingConfig holds configuration for the embedding client
type EmbeddingConfig struct {
	APIKey  string
	Model   string // default: text-embedding-3-small
	Timeout time.Duration
}

// NewEmbeddingClient creates a new embedding client
func NewEmbeddingClient(config EmbeddingConfig) *EmbeddingClient {
	if config.Model == "" {
		config.Model = "text-embedding-3-small"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &EmbeddingClient{
		apiKey:  config.APIKey,
		model:   config.Model,
		baseURL: "https://api.openai.com/v1",
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// EmbeddingRequest represents a request to the embeddings API
type EmbeddingRequest struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
}

// EmbeddingResponse represents a response from the embeddings API
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingUsage  `json:"usage"`
}

// EmbeddingData represents a single embedding in the response
type EmbeddingData struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

// EmbeddingUsage represents token usage information
type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ErrorResponse represents an error from the OpenAI API
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// CreateEmbeddings generates embeddings for the given texts
func (c *EmbeddingClient) CreateEmbeddings(ctx context.Context, texts []string) (*EmbeddingResponse, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// OpenAI recommends max 2048 texts per request, but we'll be conservative
	if len(texts) > 100 {
		return nil, fmt.Errorf("too many texts: max 100 per request, got %d", len(texts))
	}

	reqBody := EmbeddingRequest{
		Input:          texts,
		Model:          c.model,
		EncodingFormat: "float",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/embeddings", c.baseURL),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("OpenAI API error: %s (type: %s)", errResp.Error.Message, errResp.Error.Type)
	}

	var embeddingResp EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &embeddingResp, nil
}

// CreateEmbedding generates an embedding for a single text
func (c *EmbeddingClient) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	resp, err := c.CreateEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return resp.Data[0].Embedding, nil
}

// BatchEmbeddings splits texts into batches and processes them
func (c *EmbeddingClient) BatchEmbeddings(ctx context.Context, texts []string, batchSize int) ([][]float32, int, error) {
	if batchSize <= 0 || batchSize > 100 {
		batchSize = 25 // Safe default
	}

	totalEmbeddings := [][]float32{}
	totalTokens := 0

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		resp, err := c.CreateEmbeddings(ctx, batch)
		if err != nil {
			return nil, totalTokens, fmt.Errorf("failed to process batch %d-%d: %w", i, end, err)
		}

		for _, data := range resp.Data {
			totalEmbeddings = append(totalEmbeddings, data.Embedding)
		}

		totalTokens += resp.Usage.TotalTokens
	}

	return totalEmbeddings, totalTokens, nil
}
