package itemtype

import (
	"context"
	"errors"

	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
)

type Service struct {
	itemTypeRepo knowledge.ItemTypeRepository
}

func NewService(itemTypeRepo knowledge.ItemTypeRepository) *Service {
	return &Service{
		itemTypeRepo: itemTypeRepo,
	}
}

// ItemType operations

type CreateItemTypeInput struct {
	WorkspaceID common.ID
	Name        string
	Slug        common.Slug
	SchemaJSON  map[string]interface{}
	Icon        *string
	Color       *string
}

func (s *Service) CreateItemType(ctx context.Context, input CreateItemTypeInput) (*knowledge.ItemType, error) {
	if input.Name == "" {
		return nil, errors.New("name is required")
	}

	itemType := knowledge.NewItemType(
		input.WorkspaceID,
		input.Name,
		input.Slug,
		input.SchemaJSON,
	)

	if input.Icon != nil {
		itemType.SetIcon(*input.Icon)
	}

	if input.Color != nil {
		itemType.SetColor(*input.Color)
	}

	if err := s.itemTypeRepo.Create(ctx, itemType); err != nil {
		return nil, err
	}

	return itemType, nil
}

func (s *Service) GetItemType(ctx context.Context, id common.ID) (*knowledge.ItemType, error) {
	return s.itemTypeRepo.GetByID(ctx, id)
}

func (s *Service) GetItemTypeBySlug(ctx context.Context, workspaceID common.ID, slug common.Slug) (*knowledge.ItemType, error) {
	return s.itemTypeRepo.GetBySlug(ctx, workspaceID, slug)
}

func (s *Service) ListItemTypes(ctx context.Context, workspaceID common.ID) ([]*knowledge.ItemType, error) {
	return s.itemTypeRepo.ListByWorkspace(ctx, workspaceID)
}

type UpdateItemTypeInput struct {
	ID         common.ID
	Name       string
	SchemaJSON map[string]interface{}
	Icon       *string
	Color      *string
}

func (s *Service) UpdateItemType(ctx context.Context, input UpdateItemTypeInput) (*knowledge.ItemType, error) {
	itemType, err := s.itemTypeRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	itemType.Update(input.Name, input.SchemaJSON)

	if input.Icon != nil {
		itemType.SetIcon(*input.Icon)
	}

	if input.Color != nil {
		itemType.SetColor(*input.Color)
	}

	if err := s.itemTypeRepo.Update(ctx, itemType); err != nil {
		return nil, err
	}

	return itemType, nil
}

func (s *Service) DeleteItemType(ctx context.Context, id common.ID) error {
	return s.itemTypeRepo.Delete(ctx, id)
}
