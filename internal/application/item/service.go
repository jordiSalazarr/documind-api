package item

import (
	"context"
	"errors"
	"fmt"
	"log"

	"documind.jordi.org/internal/application/services"
	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
)

type Service struct {
	itemRepo        knowledge.ItemRepository
	projectRepo     knowledge.ProjectRepository
	itemTypeRepo    knowledge.ItemTypeRepository
	chunkingService *services.ChunkingService
}

func NewService(
	itemRepo knowledge.ItemRepository,
	projectRepo knowledge.ProjectRepository,
	itemTypeRepo knowledge.ItemTypeRepository,
	chunkingService *services.ChunkingService,
) *Service {
	return &Service{
		itemRepo:        itemRepo,
		projectRepo:     projectRepo,
		itemTypeRepo:    itemTypeRepo,
		chunkingService: chunkingService,
	}
}

// ===== Item Operations =====

type CreateItemInput struct {
	WorkspaceID  common.ID
	ProjectID    common.ID
	AreaID       *common.ID
	ItemTypeID   common.ID
	Title        string
	Summary      string
	BodyMd       string
	CustomFields map[string]interface{}
	Tags         []string
	OwnerUserID  common.ID
	CreatedBy    common.ID
}

func (s *Service) CreateItem(ctx context.Context, input CreateItemInput) (*knowledge.Item, *knowledge.ItemVersion, error) {
	// Validate required fields
	if input.Title == "" {
		return nil, nil, errors.New("title is required")
	}

	// Verify project exists
	projectExists, err := s.projectRepo.Exists(ctx, input.ProjectID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check project existence: %w", err)
	}
	if !projectExists {
		return nil, nil, errors.New("project not found")
	}

	// Verify item type exists
	itemTypeExists, err := s.itemTypeRepo.Exists(ctx, input.ItemTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check item type existence: %w", err)
	}
	if !itemTypeExists {
		return nil, nil, errors.New("item type not found")
	}

	// Create item
	item := knowledge.NewItem(
		input.WorkspaceID,
		input.ProjectID,
		input.ItemTypeID,
		input.OwnerUserID,
		input.CreatedBy,
	)

	if input.AreaID != nil {
		item.SetService(*input.AreaID)
	}

	if err := s.itemRepo.CreateItem(ctx, item); err != nil {
		return nil, nil, fmt.Errorf("failed to create item: %w", err)
	}

	// Create first version
	version := knowledge.NewItemVersion(
		item.ID,
		input.WorkspaceID,
		1,
		input.Title,
		input.Summary,
		input.BodyMd,
		input.CustomFields,
		input.Tags,
		input.CreatedBy,
	)

	if err := s.itemRepo.CreateItemVersion(ctx, version); err != nil {
		return nil, nil, fmt.Errorf("failed to create item version: %w", err)
	}

	// Update item's latest version
	item.IncrementVersion(input.CreatedBy)
	if err := s.itemRepo.UpdateItem(ctx, item); err != nil {
		return nil, nil, fmt.Errorf("failed to update item latest version: %w", err)
	}

	// Trigger chunking for the document (async via event bus)
	if s.chunkingService != nil {
		if err := s.chunkingService.ChunkDocument(
			item.ID,
			version.ID,
			input.WorkspaceID,
			input.Title,
			input.Summary,
			input.BodyMd,
		); err != nil {
			// Log error but don't fail the item creation
			log.Printf("Failed to chunk document for item %s: %v", item.ID, err)
		}
	}

	return item, version, nil
}

func (s *Service) GetItem(ctx context.Context, id common.ID) (*knowledge.Item, error) {
	return s.itemRepo.GetItemByID(ctx, id)
}

func (s *Service) GetItemWithLatestVersion(ctx context.Context, id common.ID) (*knowledge.Item, *knowledge.ItemVersion, error) {
	item, err := s.itemRepo.GetItemByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	version, err := s.itemRepo.GetLatestItemVersion(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return item, version, nil
}

func (s *Service) ListItemsByProject(ctx context.Context, projectID common.ID) ([]*knowledge.Item, error) {
	return s.itemRepo.ListItemsByProject(ctx, projectID)
}

func (s *Service) ListItemsByArea(ctx context.Context, serviceID common.ID) ([]*knowledge.Item, error) {
	return s.itemRepo.ListItemsByService(ctx, serviceID)
}

type UpdateItemInput struct {
	ID          common.ID
	ProjectID   *common.ID
	AreaID      *common.ID
	OwnerUserID *common.ID
	UpdatedBy   common.ID
}

func (s *Service) UpdateItem(ctx context.Context, input UpdateItemInput) (*knowledge.Item, error) {
	item, err := s.itemRepo.GetItemByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.ProjectID != nil {
		item.ProjectID = *input.ProjectID
	}

	if input.AreaID != nil {
		item.SetService(*input.AreaID)
	}

	if input.OwnerUserID != nil {
		item.OwnerUserID = *input.OwnerUserID
	}

	item.Audit.Update(input.UpdatedBy)
	item.Timestamp.Update()

	if err := s.itemRepo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *Service) PublishItem(ctx context.Context, id common.ID, updatedBy common.ID) (*knowledge.Item, error) {
	item, err := s.itemRepo.GetItemByID(ctx, id)
	if err != nil {
		return nil, err
	}

	item.Publish(updatedBy)

	if err := s.itemRepo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *Service) DeprecateItem(ctx context.Context, id common.ID, updatedBy common.ID) (*knowledge.Item, error) {
	item, err := s.itemRepo.GetItemByID(ctx, id)
	if err != nil {
		return nil, err
	}

	item.Deprecate(updatedBy)

	if err := s.itemRepo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *Service) DeleteItem(ctx context.Context, id common.ID) error {
	return s.itemRepo.DeleteItem(ctx, id)
}

// ===== ItemVersion Operations =====

type CreateItemVersionInput struct {
	ItemID       common.ID
	WorkspaceID  common.ID
	Title        string
	Summary      string
	BodyMd       string
	CustomFields map[string]interface{}
	Tags         []string
	CreatedBy    common.ID
}

func (s *Service) CreateItemVersion(ctx context.Context, input CreateItemVersionInput) (*knowledge.ItemVersion, error) {
	// Validate required fields
	if input.Title == "" {
		return nil, errors.New("title is required")
	}

	// Get item to find next version number
	item, err := s.itemRepo.GetItemByID(ctx, input.ItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	nextVersion := item.LatestVersion + 1

	// Create new version
	version := knowledge.NewItemVersion(
		input.ItemID,
		input.WorkspaceID,
		nextVersion,
		input.Title,
		input.Summary,
		input.BodyMd,
		input.CustomFields,
		input.Tags,
		input.CreatedBy,
	)

	if err := s.itemRepo.CreateItemVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("failed to create item version: %w", err)
	}

	// Update item's latest version
	item.IncrementVersion(input.CreatedBy)
	if err := s.itemRepo.UpdateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update item latest version: %w", err)
	}

	// Trigger chunking for the new version (async via event bus)
	if s.chunkingService != nil {
		if err := s.chunkingService.ChunkDocument(
			input.ItemID,
			version.ID,
			input.WorkspaceID,
			input.Title,
			input.Summary,
			input.BodyMd,
		); err != nil {
			// Log error but don't fail the version creation
			log.Printf("Failed to chunk document for version %s: %v", version.ID, err)
		}
	}

	return version, nil
}

func (s *Service) GetItemVersion(ctx context.Context, itemID common.ID, version int) (*knowledge.ItemVersion, error) {
	return s.itemRepo.GetItemVersion(ctx, itemID, version)
}

func (s *Service) GetLatestItemVersion(ctx context.Context, itemID common.ID) (*knowledge.ItemVersion, error) {
	return s.itemRepo.GetLatestItemVersion(ctx, itemID)
}

func (s *Service) ListItemVersions(ctx context.Context, itemID common.ID) ([]*knowledge.ItemVersion, error) {
	return s.itemRepo.ListItemVersions(ctx, itemID)
}

func (s *Service) PublishItemVersion(ctx context.Context, itemID common.ID, version int) (*knowledge.ItemVersion, error) {
	itemVersion, err := s.itemRepo.GetItemVersion(ctx, itemID, version)
	if err != nil {
		return nil, err
	}

	itemVersion.Publish()

	if err := s.itemRepo.UpdateItemVersionStatus(ctx, itemID, version, knowledge.ItemStatusPublished); err != nil {
		return nil, err
	}

	return itemVersion, nil
}

func (s *Service) DeprecateItemVersion(ctx context.Context, itemID common.ID, version int) (*knowledge.ItemVersion, error) {
	itemVersion, err := s.itemRepo.GetItemVersion(ctx, itemID, version)
	if err != nil {
		return nil, err
	}

	itemVersion.Deprecate()

	if err := s.itemRepo.UpdateItemVersionStatus(ctx, itemID, version, knowledge.ItemStatusDeprecated); err != nil {
		return nil, err
	}

	return itemVersion, nil
}
