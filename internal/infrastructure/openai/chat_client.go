package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ChatClient handles OpenAI chat completion API calls
type ChatClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

// ChatConfig holds configuration for the chat client
type ChatConfig struct {
	APIKey  string
	Model   string // default: gpt-4o-mini
	Timeout time.Duration
}

// NewChatClient creates a new chat completion client
func NewChatClient(config ChatConfig) *ChatClient {
	if config.Model == "" {
		config.Model = "gpt-4o-mini"
	}
	if config.Timeout == 0 {
		config.Timeout = 120 * time.Second // Longer timeout for chat completions
	}

	return &ChatClient{
		apiKey:  config.APIKey,
		model:   config.Model,
		baseURL: "https://api.openai.com/v1",
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// ChatMessage represents a message in the conversation
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, or assistant
	Content string `json:"content"`
}

// ChatRequest represents a request to the chat completions API
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatChoice represents a single completion choice
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatUsage represents token usage information
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResponse represents a response from the chat completions API
type ChatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   ChatUsage    `json:"usage"`
}

// StreamDelta represents delta content in streaming response
type StreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// StreamChoice represents a single choice in streaming response
type StreamChoice struct {
	Index        int         `json:"index"`
	Delta        StreamDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

// StreamResponse represents a streaming response chunk
type StreamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// CreateCompletion generates a chat completion (non-streaming)
func (c *ChatClient) CreateCompletion(ctx context.Context, messages []ChatMessage, maxTokens int) (*ChatResponse, error) {
	reqBody := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: 0.7,
		Stream:      false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/chat/completions", c.baseURL),
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

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &chatResp, nil
}

// StreamHandler is a callback function for handling streaming tokens
type StreamHandler func(token string, done bool) error

// CreateCompletionStream generates a chat completion with streaming
func (c *ChatClient) CreateCompletionStream(ctx context.Context, messages []ChatMessage, maxTokens int, handler StreamHandler) (*ChatUsage, error) {
	reqBody := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: 0.7,
		Stream:      true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/chat/completions", c.baseURL),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("OpenAI API error: %s (type: %s)", errResp.Error.Message, errResp.Error.Type)
	}

	reader := bufio.NewReader(resp.Body)
	var fullContent strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read stream: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip SSE comments and non-data lines
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Check for stream end
		if data == "[DONE]" {
			if err := handler("", true); err != nil {
				return nil, fmt.Errorf("handler error: %w", err)
			}
			break
		}

		var streamResp StreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			continue // Skip malformed chunks
		}

		if len(streamResp.Choices) > 0 {
			delta := streamResp.Choices[0].Delta.Content
			if delta != "" {
				fullContent.WriteString(delta)
				if err := handler(delta, false); err != nil {
					return nil, fmt.Errorf("handler error: %w", err)
				}
			}
		}
	}

	// Estimate tokens (streaming doesn't return usage, so we estimate)
	// Rough estimate: 1 token ≈ 4 characters
	estimatedTokens := len(fullContent.String()) / 4

	return &ChatUsage{
		CompletionTokens: estimatedTokens,
		TotalTokens:      estimatedTokens,
	}, nil
}

// GetModel returns the model being used
func (c *ChatClient) GetModel() string {
	return c.model
}

// BuildRAGMessages builds messages for RAG with system prompt, context, and query
func BuildRAGMessages(systemPrompt string, context string, query string, conversationHistory []ChatMessage) []ChatMessage {
	messages := []ChatMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// Add context as a system message if provided
	if context != "" {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: fmt.Sprintf("Context from knowledge base:\n\n%s", context),
		})
	}

	// Add conversation history
	messages = append(messages, conversationHistory...)

	// Add current query
	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: query,
	})

	return messages
}
