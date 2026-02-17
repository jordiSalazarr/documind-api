package domain

import (
	"context"

	shared "documind.jordi.org/internal/shared/domain"
)

// RetrievalResult represents a chunk result with scoring information
type RetrievalResult struct {
	Chunk           *Chunk
	VectorScore     float64 // Cosine similarity score (0-1)
	FullTextScore   float64 // BM25/ts_rank score
	RRFScore        float64 // Reciprocal Rank Fusion score
	RerankScore     float64 // Cross-encoder rerank score (optional)
	FinalScore      float64 // Final combined score
	VectorRank      int     // Rank in vector search results
	FullTextRank    int     // Rank in full-text search results
	RetrievalMethod string  // "vector", "fulltext", "hybrid"
}

// RetrievalConfig holds configuration for the retrieval system
type RetrievalConfig struct {
	VectorTopK      int     // Number of results from vector search (default: 50)
	FullTextTopK    int     // Number of results from full-text search (default: 50)
	RRFConstant     int     // RRF constant k (default: 60)
	FinalTopK       int     // Final number of results to return (default: 5)
	MinSimilarity   float64 // Minimum vector similarity threshold (default: 0.5)
	VectorWeight    float64 // Weight for vector search in hybrid (default: 0.7)
	FullTextWeight  float64 // Weight for full-text search in hybrid (default: 0.3)
	EnableReranking bool    // Whether to use cross-encoder reranking
	RerankTopK      int     // Number of results to rerank (default: 20)
	DiversityLambda float64 // MMR diversity parameter (default: 0.5)
	EnableMMR       bool    // Whether to apply MMR for diversity
}

// DefaultRetrievalConfig returns a default configuration
func DefaultRetrievalConfig() RetrievalConfig {
	return RetrievalConfig{
		VectorTopK:      50,
		FullTextTopK:    50,
		RRFConstant:     60,
		FinalTopK:       5,
		MinSimilarity:   0.5,
		VectorWeight:    0.7,
		FullTextWeight:  0.3,
		EnableReranking: false,
		RerankTopK:      20,
		DiversityLambda: 0.5,
		EnableMMR:       false,
	}
}

// RetrievalFilters allows filtering retrieval results
type RetrievalFilters struct {
	WorkspaceID    shared.ID
	ProjectID      *shared.ID   // Optional: filter by project
	ItemVersionIDs []shared.ID  // Optional: filter by specific item versions
	ChunkLevels    []ChunkLevel // Optional: filter by chunk levels
	IncludeCode    *bool        // Optional: include/exclude code blocks
	MaxTokens      int          // Optional: max total tokens to return
}

// RetrievalStats holds statistics about a retrieval operation
type RetrievalStats struct {
	TotalVectorResults   int
	TotalFullTextResults int
	TotalUniqueResults   int
	VectorSearchTimeMs   int64
	FullTextSearchTimeMs int64
	RerankTimeMs         int64
	TotalTimeMs          int64
}

// ItemVersionSearchResult represents an item version returned from search
type ItemVersionSearchResult struct {
	ID           shared.ID
	ItemID       shared.ID
	WorkspaceID  shared.ID
	Version      int
	Title        string
	Summary      string
	BodyMd       string
	CustomFields map[string]interface{}
	Tags         []string
	Status       string
	CreatedAt    string
	CreatedBy    shared.ID
}

// ItemSearchRepository defines the interface for item version search operations.
// This is satisfied by the knowledge context's search adapter.
type ItemSearchRepository interface {
	SearchItemVersions(ctx context.Context, workspaceID shared.ID, query string, limit, offset int) ([]*ItemVersionSearchResult, error)
	SearchItemVersionsByProject(ctx context.Context, projectID shared.ID, query string, limit, offset int) ([]*ItemVersionSearchResult, error)
	SearchByEmbedding(ctx context.Context, workspaceID shared.ID, embedding []float32, limit int, minSimilarity float64) ([]*ItemVersionSearchResult, error)
	SearchByEmbeddingInProject(ctx context.Context, projectID shared.ID, embedding []float32, limit int, minSimilarity float64) ([]*ItemVersionSearchResult, error)
}
