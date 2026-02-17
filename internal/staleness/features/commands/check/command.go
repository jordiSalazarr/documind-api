package check

import (
	"context"
	"errors"

	"documind.jordi.org/internal/shared/infrastructure/openai"
)

var (
	ErrEmptyDiff = errors.New("diff or diff_summary is required")
)

// Retriever retrieves relevant document chunks for a query.
type Retriever interface {
	Retrieve(ctx context.Context, input RetrievalInput) ([]*RetrievalResult, error)
}

// RetrievalInput contains parameters for retrieval.
type RetrievalInput struct {
	Query       string
	WorkspaceID string
	ProjectID   string // optional
	MaxResults  int
}

// RetrievalResult represents a retrieved chunk with enriched item info.
type RetrievalResult struct {
	ChunkID       string
	ItemID        string
	ItemVersionID string
	ItemTitle     string
	Content       string
	FinalScore    float64
}

// ChatClient defines the interface for LLM chat operations.
type ChatClient interface {
	CreateCompletion(ctx context.Context, messages []openai.ChatMessage, maxTokens int) (*openai.ChatResponse, error)
	GetModel() string
}

// Command contains parameters for a staleness check.
type Command struct {
	WorkspaceID   string
	Diff          string
	DiffSummary   string
	ProjectID     string // optional
	RepositoryURL string
	CommitSHA     string
}

// Result contains the staleness check results.
type Result struct {
	ReportID   string        `json:"report_id"`
	StaleCount int           `json:"stale_count"`
	Total      int           `json:"total"`
	Items      []*ItemResult `json:"items"`
}

// ItemResult represents a single item's staleness evaluation.
type ItemResult struct {
	ItemID          string  `json:"item_id"`
	ItemVersionID   string  `json:"item_version_id"`
	ItemTitle       string  `json:"item_title"`
	Stale           bool    `json:"stale"`
	Confidence      float64 `json:"confidence"`
	Explanation     string  `json:"explanation"`
	SuggestedUpdate string  `json:"suggested_update,omitempty"`
}
