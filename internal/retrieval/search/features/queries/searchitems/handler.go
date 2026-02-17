package searchitems

import (
	"context"

	searchdomain "documind.jordi.org/internal/retrieval/domain"
)

// Handler handles item version search operations
type Handler struct {
	itemRepo ItemSearchRepository
}

// NewHandler creates a new search handler
func NewHandler(itemRepo ItemSearchRepository) *Handler {
	return &Handler{
		itemRepo: itemRepo,
	}
}

// Handle performs full-text search on item versions
func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	// Set default limit if not provided
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	var versions []*searchdomain.ItemVersionSearchResult
	var err error

	// Search by project if provided, otherwise by workspace
	if q.ProjectID != nil {
		versions, err = h.itemRepo.SearchItemVersionsByProject(ctx, *q.ProjectID, q.QueryText, q.Limit, q.Offset)
	} else if q.WorkspaceID != nil {
		versions, err = h.itemRepo.SearchItemVersions(ctx, *q.WorkspaceID, q.QueryText, q.Limit, q.Offset)
	}

	if err != nil {
		return nil, err
	}

	return &Result{
		Versions:   versions,
		TotalCount: len(versions), // Note: This is just the returned count, not total matches
		Limit:      q.Limit,
		Offset:     q.Offset,
	}, nil
}

// HandleSemantic performs semantic search using vector embeddings
func (h *Handler) HandleSemantic(ctx context.Context, q SemanticQuery) (*Result, error) {
	// Set default limit if not provided
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	var versions []*searchdomain.ItemVersionSearchResult
	var err error

	// Search by project if provided, otherwise by workspace
	if q.ProjectID != nil {
		versions, err = h.itemRepo.SearchByEmbeddingInProject(ctx, *q.ProjectID, q.Embedding, q.Limit, 0.5)
	} else if q.WorkspaceID != nil {
		versions, err = h.itemRepo.SearchByEmbedding(ctx, *q.WorkspaceID, q.Embedding, q.Limit, 0.5)
	}

	if err != nil {
		return nil, err
	}

	return &Result{
		Versions:   versions,
		TotalCount: len(versions),
		Limit:      q.Limit,
		Offset:     q.Offset,
	}, nil
}
