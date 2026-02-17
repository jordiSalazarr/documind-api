package createversion

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
)

// Handler orchestrates item version creation.
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

// Handle executes the create-item-version command.
func (h *Handler) Handle(ctx context.Context, cmd Command) (*itemdomain.ItemVersion, error) {
	// Validate required fields
	if cmd.Title == "" {
		return nil, errors.New("title is required")
	}

	// Get item to find next version number
	item, err := h.repo.GetByID(ctx, cmd.ItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	nextVersion := item.LatestVersion + 1

	// Create new version
	version := itemdomain.NewItemVersion(
		cmd.ItemID,
		cmd.WorkspaceID,
		nextVersion,
		cmd.Title,
		cmd.Summary,
		cmd.BodyMd,
		cmd.CustomFields,
		cmd.Tags,
		cmd.CreatedBy,
	)

	if err := h.repo.CreateVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("failed to create item version: %w", err)
	}

	// Update item's latest version
	item.IncrementVersion(cmd.CreatedBy)
	if err := h.repo.UpdateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update item latest version: %w", err)
	}

	// Trigger chunking for the new version (async via event bus)
	if h.chunker != nil {
		if err := h.chunker.ChunkDocument(
			cmd.ItemID,
			version.ID,
			cmd.WorkspaceID,
			cmd.Title,
			cmd.Summary,
			cmd.BodyMd,
		); err != nil {
			// Log error but don't fail the version creation
			log.Printf("Failed to chunk document for version %s: %v", version.ID, err)
		}
	}

	return version, nil
}
