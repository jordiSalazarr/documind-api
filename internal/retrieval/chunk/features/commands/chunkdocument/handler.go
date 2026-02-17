package chunkdocument

import (
	"database/sql"
	"fmt"
	"log"

	"documind.jordi.org/internal/retrieval/chunker"
	shared "documind.jordi.org/internal/shared/domain"
	"documind.jordi.org/internal/shared/infrastructure/eventbus"
)

// Handler orchestrates the chunk-document use case.
type Handler struct {
	chunker   *chunker.SemanticChunker
	writeRepo WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler creates a new Handler, wiring the postgres repo internally.
func NewHandler(db *sql.DB, c *chunker.SemanticChunker, eventBus eventbus.EventBus) *Handler {
	return &Handler{
		chunker:   c,
		writeRepo: newPostgresWriteRepo(db),
		eventBus:  eventBus,
	}
}

// Handle executes the chunk document command.
func (h *Handler) Handle(cmd Command) error {
	log.Printf("Starting chunking for item %s, version %s", cmd.ItemID, cmd.VersionID)

	// Delete existing chunks for this version (in case of re-chunking)
	if err := h.writeRepo.DeleteByItemVersionID(cmd.VersionID); err != nil {
		log.Printf("Warning: Failed to delete existing chunks for version %s: %v", cmd.VersionID, err)
	}

	// Create chunks using semantic chunker
	chunks, err := h.chunker.ChunkDocument(cmd.VersionID, cmd.WorkspaceID, cmd.Title, cmd.Summary, cmd.BodyMd)
	if err != nil {
		return fmt.Errorf("failed to chunk document: %w", err)
	}

	if len(chunks) == 0 {
		log.Printf("No chunks generated for version %s", cmd.VersionID)
		return nil
	}

	log.Printf("Generated %d chunks for version %s", len(chunks), cmd.VersionID)

	// Save chunks to repository
	if err := h.writeRepo.CreateBatch(chunks); err != nil {
		return fmt.Errorf("failed to save chunks: %w", err)
	}

	log.Printf("Successfully saved %d chunks for version %s", len(chunks), cmd.VersionID)

	// Collect chunk IDs
	chunkIDs := make([]shared.ID, len(chunks))
	for i, chunk := range chunks {
		chunkIDs[i] = chunk.ID
	}

	// Publish ChunksCreated event
	chunksCreatedEvent := shared.NewChunksCreated(cmd.WorkspaceID, cmd.ItemID, cmd.VersionID, chunkIDs)
	if err := h.eventBus.Publish(chunksCreatedEvent); err != nil {
		log.Printf("Failed to publish ChunksCreated event: %v", err)
		return fmt.Errorf("failed to publish ChunksCreated event: %w", err)
	}

	log.Printf("Published ChunksCreated event for version %s", cmd.VersionID)

	// Publish ChunkEmbeddingsRequested event to trigger embedding generation
	embeddingsRequestedEvent := shared.NewChunkEmbeddingsRequested(cmd.WorkspaceID, cmd.ItemID, cmd.VersionID, chunkIDs)
	if err := h.eventBus.Publish(embeddingsRequestedEvent); err != nil {
		log.Printf("Failed to publish ChunkEmbeddingsRequested event: %v", err)
		return fmt.Errorf("failed to publish ChunkEmbeddingsRequested event: %w", err)
	}

	log.Printf("Published ChunkEmbeddingsRequested event for version %s with %d chunks", cmd.VersionID, len(chunkIDs))

	return nil
}
