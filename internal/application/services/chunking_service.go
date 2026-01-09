package services

import (
	"fmt"
	"log"

	"documind.jordi.org/internal/application/chunking"
	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/events"
	"documind.jordi.org/internal/domain/knowledge"
	"documind.jordi.org/internal/infrastructure/eventbus"
)

// ChunkingService orchestrates document chunking and event publishing
type ChunkingService struct {
	chunker   *chunking.SemanticChunker
	chunkRepo knowledge.ChunkRepository
	eventBus  eventbus.EventBus
}

// ChunkingServiceConfig holds configuration for the chunking service
type ChunkingServiceConfig struct {
	Chunker        *chunking.SemanticChunker
	ChunkRepository knowledge.ChunkRepository
	EventBus        eventbus.EventBus
}

// NewChunkingService creates a new chunking service
func NewChunkingService(config ChunkingServiceConfig) *ChunkingService {
	return &ChunkingService{
		chunker:   config.Chunker,
		chunkRepo: config.ChunkRepository,
		eventBus:  config.EventBus,
	}
}

// ChunkDocument processes a document and creates chunks
func (s *ChunkingService) ChunkDocument(
	itemID common.ID,
	versionID common.ID,
	workspaceID common.ID,
	title string,
	summary string,
	bodyMd string,
) error {
	log.Printf("Starting chunking for item %s, version %s", itemID, versionID)

	// Delete existing chunks for this version (in case of re-chunking)
	if err := s.chunkRepo.DeleteByItemVersionID(versionID); err != nil {
		log.Printf("Warning: Failed to delete existing chunks for version %s: %v", versionID, err)
		// Continue anyway - chunk creation will handle duplicates
	}

	// Create chunks using semantic chunker
	chunks, err := s.chunker.ChunkDocument(versionID, workspaceID, title, summary, bodyMd)
	if err != nil {
		return fmt.Errorf("failed to chunk document: %w", err)
	}

	if len(chunks) == 0 {
		log.Printf("No chunks generated for version %s", versionID)
		return nil
	}

	log.Printf("Generated %d chunks for version %s", len(chunks), versionID)

	// Save chunks to repository
	if err := s.chunkRepo.CreateBatch(chunks); err != nil {
		return fmt.Errorf("failed to save chunks: %w", err)
	}

	log.Printf("Successfully saved %d chunks for version %s", len(chunks), versionID)

	// Collect chunk IDs
	chunkIDs := make([]common.ID, len(chunks))
	for i, chunk := range chunks {
		chunkIDs[i] = chunk.ID
	}

	// Publish ChunksCreated event
	chunksCreatedEvent := events.NewChunksCreated(workspaceID, itemID, versionID, chunkIDs)
	if err := s.eventBus.Publish(chunksCreatedEvent); err != nil {
		log.Printf("Failed to publish ChunksCreated event: %v", err)
		return fmt.Errorf("failed to publish ChunksCreated event: %w", err)
	}

	log.Printf("Published ChunksCreated event for version %s", versionID)

	// Publish ChunkEmbeddingsRequested event to trigger embedding generation
	embeddingsRequestedEvent := events.NewChunkEmbeddingsRequested(workspaceID, itemID, versionID, chunkIDs)
	if err := s.eventBus.Publish(embeddingsRequestedEvent); err != nil {
		log.Printf("Failed to publish ChunkEmbeddingsRequested event: %v", err)
		return fmt.Errorf("failed to publish ChunkEmbeddingsRequested event: %w", err)
	}

	log.Printf("Published ChunkEmbeddingsRequested event for version %s with %d chunks", versionID, len(chunkIDs))

	return nil
}

// RechunkDocument deletes existing chunks and re-chunks the document
func (s *ChunkingService) RechunkDocument(
	itemID common.ID,
	versionID common.ID,
	workspaceID common.ID,
	title string,
	summary string,
	bodyMd string,
) error {
	log.Printf("Re-chunking document for item %s, version %s", itemID, versionID)

	// ChunkDocument already handles deletion of existing chunks
	return s.ChunkDocument(itemID, versionID, workspaceID, title, summary, bodyMd)
}

// GetChunkCount returns the number of chunks for a version
func (s *ChunkingService) GetChunkCount(versionID common.ID) (int, error) {
	return s.chunkRepo.CountByItemVersionID(versionID)
}

// GetChunks retrieves all chunks for a version
func (s *ChunkingService) GetChunks(versionID common.ID) ([]*knowledge.Chunk, error) {
	return s.chunkRepo.ListByItemVersionID(versionID)
}

// DeleteChunks deletes all chunks for a version
func (s *ChunkingService) DeleteChunks(versionID common.ID) error {
	log.Printf("Deleting chunks for version %s", versionID)

	if err := s.chunkRepo.DeleteByItemVersionID(versionID); err != nil {
		return fmt.Errorf("failed to delete chunks: %w", err)
	}

	log.Printf("Successfully deleted chunks for version %s", versionID)
	return nil
}
