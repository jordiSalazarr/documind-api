package create

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
)

// Handler orchestrates item creation.
type Handler struct {
	repo    Repository
	chunker DocumentChunker
}

// NewHandler constructs a Handler.
func NewHandler(db *sql.DB, chunker DocumentChunker) *Handler {
	return &Handler{
		repo:    newPostgresRepo(db),
		chunker: chunker,
	}
}

// Handle executes the create-item command.
func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	// Validate required fields
	if cmd.Title == "" {
		return nil, errors.New("title is required")
	}

	// Verify project exists
	projectExists, err := h.repo.ProjectExists(ctx, string(cmd.ProjectID))
	if err != nil {
		return nil, fmt.Errorf("failed to check project existence: %w", err)
	}
	if !projectExists {
		return nil, errors.New("project not found")
	}

	// Verify item type exists
	itemTypeExists, err := h.repo.ItemTypeExists(ctx, string(cmd.ItemTypeID))
	if err != nil {
		return nil, fmt.Errorf("failed to check item type existence: %w", err)
	}
	if !itemTypeExists {
		return nil, errors.New("item type not found")
	}

	// Create item
	item := itemdomain.NewItem(
		cmd.WorkspaceID,
		cmd.ProjectID,
		cmd.ItemTypeID,
		cmd.OwnerUserID,
		cmd.CreatedBy,
	)

	if cmd.ServiceID != nil {
		item.SetService(*cmd.ServiceID)
	}

	if err := h.repo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	// Create first version
	version := itemdomain.NewItemVersion(
		item.ID,
		cmd.WorkspaceID,
		1,
		cmd.Title,
		cmd.Summary,
		cmd.BodyMd,
		cmd.CustomFields,
		cmd.Tags,
		cmd.CreatedBy,
	)

	if err := h.repo.CreateItemVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("failed to create item version: %w", err)
	}

	// Update item's latest version
	item.IncrementVersion(cmd.CreatedBy)
	if err := h.repo.UpdateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update item latest version: %w", err)
	}

	// Trigger chunking for the document (async via event bus)
	if h.chunker != nil {
		if err := h.chunker.ChunkDocument(
			item.ID,
			version.ID,
			cmd.WorkspaceID,
			cmd.Title,
			cmd.Summary,
			cmd.BodyMd,
		); err != nil {
			// Log error but don't fail the item creation
			log.Printf("Failed to chunk document for item %s: %v", item.ID, err)
		}
	}

	return &Result{Item: item, Version: version}, nil
}
