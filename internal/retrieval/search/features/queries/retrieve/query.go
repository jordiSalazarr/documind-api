package retrieve

import (
	"context"

	retrievaldomain "documind.jordi.org/internal/retrieval/domain"
	"documind.jordi.org/internal/retrieval"
	shared "documind.jordi.org/internal/shared/domain"
)

// ChunkSearchRepository defines what the retrieval handler needs from the chunk read repo
type ChunkSearchRepository interface {
	SearchByEmbedding(workspaceID shared.ID, embedding []float32, limit int, minSimilarity float64) ([]*retrievaldomain.Chunk, error)
	SearchByText(workspaceID shared.ID, query string, limit int) ([]*retrievaldomain.Chunk, error)
}

// Reranker scores and re-orders retrieval results by semantic relevance to the query.
type Reranker interface {
	Rerank(ctx context.Context, query string, results []*retrievaldomain.RetrievalResult, topK int) ([]*retrievaldomain.RetrievalResult, error)
}

// Input contains parameters for hybrid retrieval
type Input struct {
	Query             string
	WorkspaceID       shared.ID
	ProjectID         *shared.ID
	MaxResults        int
	MaxTokens         int  // Max total tokens across all results
	EnableQueryExpand bool // Enable query expansion (default: true)
	EnableMultiQuery  bool // Enable multi-query retrieval for better recall
}

// Output contains the retrieval results
type Output struct {
	Results        []*retrievaldomain.RetrievalResult
	Stats          *retrievaldomain.RetrievalStats
	ProcessedQuery *retrieval.ProcessedQuery // Query processing information
}
