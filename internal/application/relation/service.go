package relation

import (
	"context"
	"fmt"

	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
)

type Service struct {
	relationTypeRepo knowledge.RelationTypeRepository
	relationRepo     knowledge.RelationRepository
}

func NewService(
	relationTypeRepo knowledge.RelationTypeRepository,
	relationRepo knowledge.RelationRepository,
) *Service {
	return &Service{
		relationTypeRepo: relationTypeRepo,
		relationRepo:     relationRepo,
	}
}

// ===== Relation Type Operations =====

type CreateRelationTypeInput struct {
	WorkspaceID   common.ID
	Name          string
	Slug          common.Slug
	IsDirectional bool
}

func (s *Service) CreateRelationType(ctx context.Context, input CreateRelationTypeInput) (*knowledge.RelationType, error) {
	// Check if slug already exists in workspace
	existing, err := s.relationTypeRepo.GetBySlug(ctx, input.WorkspaceID, input.Slug)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("relation type with slug '%s' already exists", input.Slug)
	}

	relationType := knowledge.NewRelationType(
		input.WorkspaceID,
		input.Name,
		input.Slug,
		input.IsDirectional,
	)

	if err := s.relationTypeRepo.Create(ctx, relationType); err != nil {
		return nil, fmt.Errorf("failed to create relation type: %w", err)
	}

	return relationType, nil
}

func (s *Service) GetRelationType(ctx context.Context, id common.ID) (*knowledge.RelationType, error) {
	return s.relationTypeRepo.GetByID(ctx, id)
}

func (s *Service) GetRelationTypeBySlug(ctx context.Context, workspaceID common.ID, slug common.Slug) (*knowledge.RelationType, error) {
	return s.relationTypeRepo.GetBySlug(ctx, workspaceID, slug)
}

func (s *Service) ListRelationTypes(ctx context.Context, workspaceID common.ID) ([]*knowledge.RelationType, error) {
	return s.relationTypeRepo.ListByWorkspace(ctx, workspaceID)
}

type UpdateRelationTypeInput struct {
	ID            common.ID
	Name          *string
	Slug          *common.Slug
	IsDirectional *bool
}

func (s *Service) UpdateRelationType(ctx context.Context, input UpdateRelationTypeInput) (*knowledge.RelationType, error) {
	relationType, err := s.relationTypeRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		relationType.Name = *input.Name
	}

	if input.Slug != nil {
		// Check if new slug already exists
		existing, err := s.relationTypeRepo.GetBySlug(ctx, relationType.WorkspaceID, *input.Slug)
		if err == nil && existing != nil && existing.ID != relationType.ID {
			return nil, fmt.Errorf("relation type with slug '%s' already exists", *input.Slug)
		}
		relationType.Slug = *input.Slug
	}

	if input.IsDirectional != nil {
		relationType.IsDirectional = *input.IsDirectional
	}

	relationType.Update()

	if err := s.relationTypeRepo.Update(ctx, relationType); err != nil {
		return nil, fmt.Errorf("failed to update relation type: %w", err)
	}

	return relationType, nil
}

func (s *Service) DeleteRelationType(ctx context.Context, id common.ID) error {
	return s.relationTypeRepo.Delete(ctx, id)
}

// ===== Relation Operations =====

type CreateRelationInput struct {
	WorkspaceID    common.ID
	FromItemID     common.ID
	ToItemID       common.ID
	RelationTypeID common.ID
	CreatedBy      common.ID
}

func (s *Service) CreateRelation(ctx context.Context, input CreateRelationInput) (*knowledge.Relation, error) {
	// Verify relation type exists
	_, err := s.relationTypeRepo.GetByID(ctx, input.RelationTypeID)
	if err != nil {
		return nil, fmt.Errorf("relation type not found: %w", err)
	}

	// Check if relation already exists
	existing, err := s.relationRepo.GetByItems(ctx, input.FromItemID, input.ToItemID, input.RelationTypeID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("relation already exists between these items")
	}

	relation, err := knowledge.NewRelation(
		input.WorkspaceID,
		input.FromItemID,
		input.ToItemID,
		input.RelationTypeID,
		input.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create relation: %w", err)
	}

	if err := s.relationRepo.Create(ctx, relation); err != nil {
		return nil, fmt.Errorf("failed to save relation: %w", err)
	}

	// Fetch the relation with type information
	return s.relationRepo.GetByID(ctx, relation.ID)
}

func (s *Service) GetRelation(ctx context.Context, id common.ID) (*knowledge.Relation, error) {
	return s.relationRepo.GetByID(ctx, id)
}

func (s *Service) ListRelationsByFromItem(ctx context.Context, fromItemID common.ID) ([]*knowledge.Relation, error) {
	return s.relationRepo.ListByFromItem(ctx, fromItemID)
}

func (s *Service) ListRelationsByToItem(ctx context.Context, toItemID common.ID) ([]*knowledge.Relation, error) {
	return s.relationRepo.ListByToItem(ctx, toItemID)
}

func (s *Service) ListRelationsByItem(ctx context.Context, itemID common.ID) ([]*knowledge.Relation, error) {
	return s.relationRepo.ListByItem(ctx, itemID)
}

func (s *Service) ListRelationsByType(ctx context.Context, relationTypeID common.ID) ([]*knowledge.Relation, error) {
	return s.relationRepo.ListByRelationType(ctx, relationTypeID)
}

func (s *Service) DeleteRelation(ctx context.Context, id common.ID) error {
	relation, err := s.relationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	relation.SoftDelete()
	return s.relationRepo.Delete(ctx, id)
}
