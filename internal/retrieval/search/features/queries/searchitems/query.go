package searchitems

import (
	"context"

	searchdomain "documind.jordi.org/internal/retrieval/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

// ItemSearchRepository defines the interface for item version search operations.
// This is satisfied by the item module's postgres repository.
type ItemSearchRepository interface {
	SearchItemVersions(ctx context.Context, workspaceID shared.ID, query string, limit, offset int) ([]*searchdomain.ItemVersionSearchResult, error)
	SearchItemVersionsByProject(ctx context.Context, projectID shared.ID, query string, limit, offset int) ([]*searchdomain.ItemVersionSearchResult, error)
	SearchByEmbedding(ctx context.Context, workspaceID shared.ID, embedding []float32, limit int, minSimilarity float64) ([]*searchdomain.ItemVersionSearchResult, error)
	SearchByEmbeddingInProject(ctx context.Context, projectID shared.ID, embedding []float32, limit int, minSimilarity float64) ([]*searchdomain.ItemVersionSearchResult, error)
}

// Query contains parameters for item version search
type Query struct {
	QueryText   string
	WorkspaceID *shared.ID
	ProjectID   *shared.ID
	Limit       int
	Offset      int
}

// Result contains search results with pagination info
type Result struct {
	Versions   []*searchdomain.ItemVersionSearchResult
	TotalCount int
	Limit      int
	Offset     int
}

// SemanticQuery contains parameters for semantic search
type SemanticQuery struct {
	Embedding   []float32
	WorkspaceID *shared.ID
	ProjectID   *shared.ID
	Limit       int
	Offset      int
}
