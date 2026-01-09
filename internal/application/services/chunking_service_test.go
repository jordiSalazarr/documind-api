package services

import (
	"errors"
	"testing"

	"documind.jordi.org/internal/application/chunking"
	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/events"
	"documind.jordi.org/internal/domain/knowledge"
	"documind.jordi.org/internal/infrastructure/eventbus"
)

// Mock implementations for testing

type mockChunkRepository struct {
	chunks              []*knowledge.Chunk
	createBatchErr      error
	deleteErr           error
	listErr             error
	countByVersionID    map[common.ID]int
}

func (m *mockChunkRepository) Create(chunk *knowledge.Chunk) error {
	m.chunks = append(m.chunks, chunk)
	return nil
}

func (m *mockChunkRepository) CreateBatch(chunks []*knowledge.Chunk) error {
	if m.createBatchErr != nil {
		return m.createBatchErr
	}
	m.chunks = append(m.chunks, chunks...)
	return nil
}

func (m *mockChunkRepository) GetByID(id common.ID) (*knowledge.Chunk, error) {
	for _, chunk := range m.chunks {
		if chunk.ID == id {
			return chunk, nil
		}
	}
	return nil, errors.New("chunk not found")
}

func (m *mockChunkRepository) ListByItemVersionID(itemVersionID common.ID) ([]*knowledge.Chunk, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	result := []*knowledge.Chunk{}
	for _, chunk := range m.chunks {
		if chunk.ItemVersionID == itemVersionID {
			result = append(result, chunk)
		}
	}
	return result, nil
}

func (m *mockChunkRepository) UpdateEmbedding(id common.ID, embedding []float32) error {
	return nil
}

func (m *mockChunkRepository) UpdateEmbeddingsBatch(updates map[common.ID][]float32) error {
	return nil
}

func (m *mockChunkRepository) SearchByEmbedding(workspaceID common.ID, embedding []float32, limit int, minSimilarity float64) ([]*knowledge.Chunk, error) {
	return nil, nil
}

func (m *mockChunkRepository) SearchByText(workspaceID common.ID, query string, limit int) ([]*knowledge.Chunk, error) {
	return nil, nil
}

func (m *mockChunkRepository) DeleteByItemVersionID(itemVersionID common.ID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	filtered := []*knowledge.Chunk{}
	for _, chunk := range m.chunks {
		if chunk.ItemVersionID != itemVersionID {
			filtered = append(filtered, chunk)
		}
	}
	m.chunks = filtered
	return nil
}

func (m *mockChunkRepository) CountByItemVersionID(itemVersionID common.ID) (int, error) {
	if count, ok := m.countByVersionID[itemVersionID]; ok {
		return count, nil
	}
	count := 0
	for _, chunk := range m.chunks {
		if chunk.ItemVersionID == itemVersionID {
			count++
		}
	}
	return count, nil
}

type mockEventBus struct {
	publishedEvents []events.Event
	publishErr      error
}

func (m *mockEventBus) Publish(event events.Event) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

func (m *mockEventBus) Subscribe(eventType events.EventType, handler eventbus.EventHandler) error {
	return nil
}

func (m *mockEventBus) Start() error {
	return nil
}

func (m *mockEventBus) Stop() error {
	return nil
}

func TestChunkingService_ChunkDocument_Success(t *testing.T) {
	mockRepo := &mockChunkRepository{
		chunks:           []*knowledge.Chunk{},
		countByVersionID: make(map[common.ID]int),
	}
	mockBus := &mockEventBus{
		publishedEvents: []events.Event{},
	}

	chunker := chunking.NewSemanticChunker(chunking.DefaultChunkingConfig())
	service := NewChunkingService(ChunkingServiceConfig{
		Chunker:         chunker,
		ChunkRepository: mockRepo,
		EventBus:        mockBus,
	})

	itemID := common.NewID()
	versionID := common.NewID()
	workspaceID := common.NewID()

	err := service.ChunkDocument(
		itemID,
		versionID,
		workspaceID,
		"Test Document",
		"A test summary",
		"This is the body content of the document.",
	)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify chunks were created
	if len(mockRepo.chunks) == 0 {
		t.Error("Expected chunks to be created")
	}

	// Verify events were published
	if len(mockBus.publishedEvents) != 2 {
		t.Errorf("Expected 2 events to be published, got %d", len(mockBus.publishedEvents))
	}

	// Verify ChunksCreated event
	chunksCreatedFound := false
	for _, event := range mockBus.publishedEvents {
		if event.GetType() == events.ChunksCreatedEvent {
			chunksCreatedFound = true
			chunksCreatedEvent := event.(*events.ChunksCreated)
			if chunksCreatedEvent.VersionID != versionID.String() {
				t.Errorf("Expected VersionID %s, got %s", versionID, chunksCreatedEvent.VersionID)
			}
		}
	}
	if !chunksCreatedFound {
		t.Error("Expected ChunksCreated event to be published")
	}

	// Verify ChunkEmbeddingsRequested event
	embeddingsRequestedFound := false
	for _, event := range mockBus.publishedEvents {
		if event.GetType() == events.ChunkEmbeddingsRequestedEvent {
			embeddingsRequestedFound = true
		}
	}
	if !embeddingsRequestedFound {
		t.Error("Expected ChunkEmbeddingsRequested event to be published")
	}
}

func TestChunkingService_ChunkDocument_EmptyContent(t *testing.T) {
	mockRepo := &mockChunkRepository{
		chunks:           []*knowledge.Chunk{},
		countByVersionID: make(map[common.ID]int),
	}
	mockBus := &mockEventBus{
		publishedEvents: []events.Event{},
	}

	chunker := chunking.NewSemanticChunker(chunking.DefaultChunkingConfig())
	service := NewChunkingService(ChunkingServiceConfig{
		Chunker:         chunker,
		ChunkRepository: mockRepo,
		EventBus:        mockBus,
	})

	itemID := common.NewID()
	versionID := common.NewID()
	workspaceID := common.NewID()

	err := service.ChunkDocument(itemID, versionID, workspaceID, "", "", "")

	if err != nil {
		t.Fatalf("Expected no error for empty content, got: %v", err)
	}

	// No chunks should be created for empty content
	if len(mockRepo.chunks) != 0 {
		t.Errorf("Expected 0 chunks for empty content, got %d", len(mockRepo.chunks))
	}

	// No events should be published for empty content
	if len(mockBus.publishedEvents) != 0 {
		t.Errorf("Expected 0 events for empty content, got %d", len(mockBus.publishedEvents))
	}
}

func TestChunkingService_ChunkDocument_RepositoryError(t *testing.T) {
	mockRepo := &mockChunkRepository{
		chunks:           []*knowledge.Chunk{},
		createBatchErr:   errors.New("database error"),
		countByVersionID: make(map[common.ID]int),
	}
	mockBus := &mockEventBus{
		publishedEvents: []events.Event{},
	}

	chunker := chunking.NewSemanticChunker(chunking.DefaultChunkingConfig())
	service := NewChunkingService(ChunkingServiceConfig{
		Chunker:         chunker,
		ChunkRepository: mockRepo,
		EventBus:        mockBus,
	})

	itemID := common.NewID()
	versionID := common.NewID()
	workspaceID := common.NewID()

	err := service.ChunkDocument(
		itemID,
		versionID,
		workspaceID,
		"Test",
		"",
		"Content",
	)

	if err == nil {
		t.Fatal("Expected error when repository fails")
	}

	// Events should not be published when chunking fails
	if len(mockBus.publishedEvents) != 0 {
		t.Errorf("Expected 0 events when chunking fails, got %d", len(mockBus.publishedEvents))
	}
}

func TestChunkingService_GetChunkCount(t *testing.T) {
	versionID := common.NewID()
	mockRepo := &mockChunkRepository{
		chunks: []*knowledge.Chunk{},
		countByVersionID: map[common.ID]int{
			versionID: 5,
		},
	}
	mockBus := &mockEventBus{}

	chunker := chunking.NewSemanticChunker(chunking.DefaultChunkingConfig())
	service := NewChunkingService(ChunkingServiceConfig{
		Chunker:         chunker,
		ChunkRepository: mockRepo,
		EventBus:        mockBus,
	})

	count, err := service.GetChunkCount(versionID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}
}

func TestChunkingService_GetChunks(t *testing.T) {
	versionID := common.NewID()
	workspaceID := common.NewID()

	chunk1 := knowledge.NewChunk(versionID, workspaceID, "Content 1", 0, knowledge.ChunkLevelParagraph, 10)
	chunk2 := knowledge.NewChunk(versionID, workspaceID, "Content 2", 1, knowledge.ChunkLevelParagraph, 10)

	mockRepo := &mockChunkRepository{
		chunks:           []*knowledge.Chunk{chunk1, chunk2},
		countByVersionID: make(map[common.ID]int),
	}
	mockBus := &mockEventBus{}

	chunker := chunking.NewSemanticChunker(chunking.DefaultChunkingConfig())
	service := NewChunkingService(ChunkingServiceConfig{
		Chunker:         chunker,
		ChunkRepository: mockRepo,
		EventBus:        mockBus,
	})

	chunks, err := service.GetChunks(versionID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %d", len(chunks))
	}
}

func TestChunkingService_DeleteChunks(t *testing.T) {
	versionID := common.NewID()
	workspaceID := common.NewID()

	chunk1 := knowledge.NewChunk(versionID, workspaceID, "Content 1", 0, knowledge.ChunkLevelParagraph, 10)

	mockRepo := &mockChunkRepository{
		chunks:           []*knowledge.Chunk{chunk1},
		countByVersionID: make(map[common.ID]int),
	}
	mockBus := &mockEventBus{}

	chunker := chunking.NewSemanticChunker(chunking.DefaultChunkingConfig())
	service := NewChunkingService(ChunkingServiceConfig{
		Chunker:         chunker,
		ChunkRepository: mockRepo,
		EventBus:        mockBus,
	})

	err := service.DeleteChunks(versionID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify chunks were deleted
	if len(mockRepo.chunks) != 0 {
		t.Errorf("Expected 0 chunks after deletion, got %d", len(mockRepo.chunks))
	}
}

func TestChunkingService_RechunkDocument(t *testing.T) {
	mockRepo := &mockChunkRepository{
		chunks:           []*knowledge.Chunk{},
		countByVersionID: make(map[common.ID]int),
	}
	mockBus := &mockEventBus{
		publishedEvents: []events.Event{},
	}

	chunker := chunking.NewSemanticChunker(chunking.DefaultChunkingConfig())
	service := NewChunkingService(ChunkingServiceConfig{
		Chunker:         chunker,
		ChunkRepository: mockRepo,
		EventBus:        mockBus,
	})

	itemID := common.NewID()
	versionID := common.NewID()
	workspaceID := common.NewID()

	// Add some existing chunks
	existingChunk := knowledge.NewChunk(versionID, workspaceID, "Old content", 0, knowledge.ChunkLevelParagraph, 10)
	mockRepo.chunks = append(mockRepo.chunks, existingChunk)

	err := service.RechunkDocument(
		itemID,
		versionID,
		workspaceID,
		"Updated Document",
		"Updated summary",
		"Updated body content.",
	)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify new chunks were created (old ones deleted)
	if len(mockRepo.chunks) == 0 {
		t.Error("Expected new chunks to be created")
	}

	// Verify events were published
	if len(mockBus.publishedEvents) < 2 {
		t.Errorf("Expected at least 2 events to be published, got %d", len(mockBus.publishedEvents))
	}
}
