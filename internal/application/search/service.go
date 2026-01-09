package search

import (
	"context"

	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
)

// ItemRepository interface with search methods
type ItemRepository interface {
	SearchItemVersions(ctx context.Context, workspaceID common.ID, query string, limit, offset int) ([]*knowledge.ItemVersion, error)
	SearchItemVersionsByProject(ctx context.Context, projectID common.ID, query string, limit, offset int) ([]*knowledge.ItemVersion, error)
	SearchItemVersionsByEmbedding(ctx context.Context, workspaceID common.ID, embedding []float32, limit, offset int) ([]*knowledge.ItemVersion, error)
	SearchItemVersionsByEmbeddingInProject(ctx context.Context, projectID common.ID, embedding []float32, limit, offset int) ([]*knowledge.ItemVersion, error)
	UpdateItemVersionEmbedding(ctx context.Context, versionID common.ID, embedding []float32) error
}

type Service struct {
	itemRepo ItemRepository
}

func NewService(itemRepo ItemRepository) *Service {
	return &Service{
		itemRepo: itemRepo,
	}
}

type SearchInput struct {
	Query       string
	WorkspaceID *common.ID
	ProjectID   *common.ID
	Limit       int
	Offset      int
}

type SearchResult struct {
	Versions   []*knowledge.ItemVersion
	TotalCount int
	Limit      int
	Offset     int
}

func (s *Service) SearchItemVersions(ctx context.Context, input SearchInput) (*SearchResult, error) {
	// Set default limit if not provided
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	var versions []*knowledge.ItemVersion
	var err error

	// Search by project if provided, otherwise by workspace
	if input.ProjectID != nil {
		versions, err = s.itemRepo.SearchItemVersionsByProject(ctx, *input.ProjectID, input.Query, input.Limit, input.Offset)
	} else if input.WorkspaceID != nil {
		versions, err = s.itemRepo.SearchItemVersions(ctx, *input.WorkspaceID, input.Query, input.Limit, input.Offset)
	}

	if err != nil {
		return nil, err
	}

	return &SearchResult{
		Versions:   versions,
		TotalCount: len(versions), // Note: This is just the returned count, not total matches
		Limit:      input.Limit,
		Offset:     input.Offset,
	}, nil
}

// SemanticSearchInput contains parameters for semantic search
type SemanticSearchInput struct {
	Embedding   []float32
	WorkspaceID *common.ID
	ProjectID   *common.ID
	Limit       int
	Offset      int
}

// SearchItemVersionsBySemantic performs semantic search using vector embeddings
func (s *Service) SearchItemVersionsBySemantic(ctx context.Context, input SemanticSearchInput) (*SearchResult, error) {
	// Set default limit if not provided
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	var versions []*knowledge.ItemVersion
	var err error

	// Search by project if provided, otherwise by workspace
	if input.ProjectID != nil {
		versions, err = s.itemRepo.SearchItemVersionsByEmbeddingInProject(ctx, *input.ProjectID, input.Embedding, input.Limit, input.Offset)
	} else if input.WorkspaceID != nil {
		versions, err = s.itemRepo.SearchItemVersionsByEmbedding(ctx, *input.WorkspaceID, input.Embedding, input.Limit, input.Offset)
	}

	if err != nil {
		return nil, err
	}

	return &SearchResult{
		Versions:   versions,
		TotalCount: len(versions),
		Limit:      input.Limit,
		Offset:     input.Offset,
	}, nil
}

// UpdateEmbedding updates the embedding for an item version
func (s *Service) UpdateEmbedding(ctx context.Context, versionID common.ID, embedding []float32) error {
	return s.itemRepo.UpdateItemVersionEmbedding(ctx, versionID, embedding)
}
