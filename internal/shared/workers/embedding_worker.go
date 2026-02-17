package workers

import (
	"context"
	"fmt"
	"log"

	"documind.jordi.org/internal/shared/domain"
	"documind.jordi.org/internal/shared/infrastructure/eventbus"
	"documind.jordi.org/internal/shared/infrastructure/openai"
)

// ChunkForEmbedding represents the minimal chunk data needed by the embedding worker
type ChunkForEmbedding struct {
	ID        domain.ID
	Content   string
	Embedding []float32
}

// ChunkEmbeddingRepository defines the repository interface needed by the embedding worker
type ChunkEmbeddingRepository interface {
	ListByItemVersionID(versionID domain.ID) ([]*ChunkForEmbedding, error)
	UpdateEmbeddingsBatch(updates map[domain.ID][]float32) error
}

// EmbeddingWorker processes chunk embedding requests
type EmbeddingWorker struct {
	chunkRepo       ChunkEmbeddingRepository
	embeddingClient *openai.EmbeddingClient
	eventBus        eventbus.EventBus
	batchSize       int
}

// EmbeddingWorkerConfig holds configuration for the embedding worker
type EmbeddingWorkerConfig struct {
	ChunkRepository ChunkEmbeddingRepository
	EmbeddingClient *openai.EmbeddingClient
	EventBus        eventbus.EventBus
	BatchSize       int // Number of chunks to process in a single OpenAI API call
}

// NewEmbeddingWorker creates a new embedding worker
func NewEmbeddingWorker(config EmbeddingWorkerConfig) *EmbeddingWorker {
	if config.BatchSize <= 0 {
		config.BatchSize = 25 // Safe default for OpenAI batch processing
	}

	return &EmbeddingWorker{
		chunkRepo:       config.ChunkRepository,
		embeddingClient: config.EmbeddingClient,
		eventBus:        config.EventBus,
		batchSize:       config.BatchSize,
	}
}

// Start registers event handlers and starts processing events
func (w *EmbeddingWorker) Start() error {
	log.Println("Starting embedding worker...")

	// Subscribe to chunk embedding requests
	if err := w.eventBus.Subscribe(domain.ChunkEmbeddingsRequestedEvent, w.handleChunkEmbeddingsRequested); err != nil {
		return fmt.Errorf("failed to subscribe to ChunkEmbeddingsRequested: %w", err)
	}

	log.Println("Embedding worker started successfully")
	return nil
}

// handleChunkEmbeddingsRequested processes chunk embedding requests
func (w *EmbeddingWorker) handleChunkEmbeddingsRequested(event domain.Event) error {
	embeddingEvent, ok := event.(*domain.ChunkEmbeddingsRequested)
	if !ok {
		return fmt.Errorf("invalid event type: expected ChunkEmbeddingsRequested")
	}

	log.Printf("Processing chunk embeddings request for version %s (%d chunks)",
		embeddingEvent.VersionID, embeddingEvent.ChunkCount)

	ctx := context.Background()
	versionID := domain.ID(embeddingEvent.VersionID)

	// Fetch all chunks for this version
	chunks, err := w.chunkRepo.ListByItemVersionID(versionID)
	if err != nil {
		return fmt.Errorf("failed to fetch chunks: %w", err)
	}

	if len(chunks) == 0 {
		log.Printf("No chunks found for version %s", embeddingEvent.VersionID)
		return nil
	}

	log.Printf("Fetched %d chunks for embedding generation", len(chunks))

	// Filter out chunks that already have embeddings
	chunksNeedingEmbeddings := make([]*ChunkForEmbedding, 0, len(chunks))
	for _, chunk := range chunks {
		if len(chunk.Embedding) == 0 {
			chunksNeedingEmbeddings = append(chunksNeedingEmbeddings, chunk)
		}
	}

	if len(chunksNeedingEmbeddings) == 0 {
		log.Printf("All chunks already have embeddings for version %s", embeddingEvent.VersionID)
		return nil
	}

	log.Printf("Generating embeddings for %d chunks (skipped %d with existing embeddings)",
		len(chunksNeedingEmbeddings), len(chunks)-len(chunksNeedingEmbeddings))

	// Extract text content from chunks
	texts := make([]string, len(chunksNeedingEmbeddings))
	for i, chunk := range chunksNeedingEmbeddings {
		texts[i] = chunk.Content
	}

	// Generate embeddings in batches
	embeddings, tokensUsed, err := w.embeddingClient.BatchEmbeddings(ctx, texts, w.batchSize)
	if err != nil {
		log.Printf("Failed to generate embeddings for version %s: %v", embeddingEvent.VersionID, err)

		// Publish failure event
		workspaceID := domain.ID(embeddingEvent.WorkspaceID)
		failureEvent := domain.NewChunkEmbeddingsGenerated(
			workspaceID,
			versionID,
			len(chunksNeedingEmbeddings),
			0,
			0,
			len(chunksNeedingEmbeddings),
		)
		if publishErr := w.eventBus.Publish(failureEvent); publishErr != nil {
			log.Printf("Failed to publish failure event: %v", publishErr)
		}

		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Update chunks with generated embeddings
	updates := make(map[domain.ID][]float32)
	for i, chunk := range chunksNeedingEmbeddings {
		updates[chunk.ID] = embeddings[i]
	}

	if err := w.chunkRepo.UpdateEmbeddingsBatch(updates); err != nil {
		log.Printf("Failed to save embeddings for version %s: %v", embeddingEvent.VersionID, err)

		// Publish failure event
		workspaceID := domain.ID(embeddingEvent.WorkspaceID)
		failureEvent := domain.NewChunkEmbeddingsGenerated(
			workspaceID,
			versionID,
			len(chunksNeedingEmbeddings),
			tokensUsed,
			0,
			len(chunksNeedingEmbeddings),
		)
		if publishErr := w.eventBus.Publish(failureEvent); publishErr != nil {
			log.Printf("Failed to publish failure event: %v", publishErr)
		}

		return fmt.Errorf("failed to save embeddings: %w", err)
	}

	log.Printf("Successfully generated and saved embeddings for %d chunks (tokens used: %d)",
		len(chunksNeedingEmbeddings), tokensUsed)

	// Publish success event
	workspaceID := domain.ID(embeddingEvent.WorkspaceID)
	successEvent := domain.NewChunkEmbeddingsGenerated(
		workspaceID,
		versionID,
		len(chunksNeedingEmbeddings),
		tokensUsed,
		len(chunksNeedingEmbeddings),
		0,
	)

	if err := w.eventBus.Publish(successEvent); err != nil {
		log.Printf("Failed to publish success event: %v", err)
		return err
	}

	return nil
}
