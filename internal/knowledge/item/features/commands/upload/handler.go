package upload

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
)

// Handler orchestrates file upload → parse → create item → chunk.
type Handler struct {
	repo      Repository
	parser    FileParser
	validator FileValidator
	chunker   DocumentChunker
}

// NewHandler constructs an upload Handler.
func NewHandler(db *sql.DB, parser FileParser, validator FileValidator, chunker DocumentChunker) *Handler {
	return &Handler{
		repo:      newPostgresRepo(db),
		parser:    parser,
		validator: validator,
		chunker:   chunker,
	}
}

// Handle executes the file upload command.
func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	// 1. Validate file
	if err := h.validator.Validate(cmd.Filename, cmd.FileSize); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// 2. Verify project exists
	projectExists, err := h.repo.ProjectExists(ctx, string(cmd.ProjectID))
	if err != nil {
		return nil, fmt.Errorf("failed to check project: %w", err)
	}
	if !projectExists {
		return nil, errors.New("project not found")
	}

	// 3. Verify item type exists
	itemTypeExists, err := h.repo.ItemTypeExists(ctx, string(cmd.ItemTypeID))
	if err != nil {
		return nil, fmt.Errorf("failed to check item type: %w", err)
	}
	if !itemTypeExists {
		return nil, errors.New("item type not found")
	}

	// 4. Parse file into structured document
	parsed, err := h.parser.Parse(ctx, cmd.Filename, cmd.FileReader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// 5. Check for exact duplicate via content hash
	hash := sha256.Sum256([]byte(parsed.Content))
	contentHash := hex.EncodeToString(hash[:])

	existingItemID, existingVersionID, found, err := h.repo.FindByContentHash(ctx, string(cmd.WorkspaceID), contentHash)
	if err != nil {
		log.Printf("Dedup check failed (continuing): %v", err)
	}
	if found {
		return nil, fmt.Errorf(
			"duplicate content: this file matches existing item %s (version %s) in this workspace",
			existingItemID, existingVersionID,
		)
	}

	// 6. Create Item (status=draft)
	item := itemdomain.NewItem(
		cmd.WorkspaceID,
		cmd.ProjectID,
		cmd.ItemTypeID,
		cmd.OwnerUserID,
		cmd.CreatedBy,
	)

	if cmd.AreaID != nil {
		item.SetService(*cmd.AreaID)
	}

	if err := h.repo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	// 7. Create ItemVersion (version=1, body_md = parsed content, content_hash for dedup)
	customFields := map[string]interface{}{
		"source_format": parsed.Metadata.SourceFormat,
		"word_count":    parsed.Metadata.WordCount,
		"page_count":    parsed.Metadata.PageCount,
		"original_file": cmd.Filename,
	}
	if parsed.Metadata.Author != "" {
		customFields["author"] = parsed.Metadata.Author
	}
	if parsed.Metadata.CodeLanguage != "" {
		customFields["code_language"] = parsed.Metadata.CodeLanguage
	}

	version := itemdomain.NewItemVersion(
		item.ID,
		cmd.WorkspaceID,
		1,
		parsed.Title,
		"", // summary — could be generated later
		parsed.Content,
		customFields,
		nil, // tags
		cmd.CreatedBy,
	)

	if err := h.repo.CreateItemVersionWithHash(ctx, version, contentHash); err != nil {
		return nil, fmt.Errorf("failed to create item version: %w", err)
	}

	// 8. Update item's latest version
	item.IncrementVersion(cmd.CreatedBy)
	if err := h.repo.UpdateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update item latest version: %w", err)
	}

	// 9. Trigger chunking (async via event bus)
	if h.chunker != nil {
		if err := h.chunker.ChunkDocument(
			item.ID,
			version.ID,
			cmd.WorkspaceID,
			parsed.Title,
			"",
			parsed.Content,
		); err != nil {
			log.Printf("Failed to chunk uploaded document for item %s: %v", item.ID, err)
		}
	}

	return &Result{
		Item:    item,
		Version: version,
		Parsing: ParsingResult{
			SourceFormat: parsed.Metadata.SourceFormat,
			PageCount:    parsed.Metadata.PageCount,
			WordCount:    parsed.Metadata.WordCount,
			Status:       "chunking_in_progress",
		},
	}, nil
}
