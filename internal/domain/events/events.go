package events

import (
	"documind.jordi.org/internal/domain/common"
	"time"
)

// EventType represents the type of domain event
type EventType string

const (
	ItemCreatedEvent               EventType = "ItemCreated"
	VersionCreatedEvent            EventType = "VersionCreated"
	PreviousVersionDeprecatedEvent EventType = "PreviousVersionDeprecated"
	ItemPublishedEvent             EventType = "ItemPublished"
	SourceLinkedToItemEvent        EventType = "SourceLinkedToItem"
	RelationCreatedEvent           EventType = "RelationCreated"
	EmbeddingGeneratedEvent        EventType = "EmbeddingGenerated"
	EmbeddingRequestedEvent        EventType = "EmbeddingRequested"
	ChunksCreatedEvent             EventType = "ChunksCreated"
	ChunkEmbeddingsRequestedEvent  EventType = "ChunkEmbeddingsRequested"
	ChunkEmbeddingsGeneratedEvent  EventType = "ChunkEmbeddingsGenerated"
)

// Event represents a domain event
type Event interface {
	GetType() EventType
	GetID() string
	GetTimestamp() time.Time
	GetWorkspaceID() string
}

// BaseEvent contains common event fields
type BaseEvent struct {
	ID          string
	Type        EventType
	WorkspaceID string
	Timestamp   time.Time
}

func (e BaseEvent) GetType() EventType {
	return e.Type
}

func (e BaseEvent) GetID() string {
	return e.ID
}

func (e BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e BaseEvent) GetWorkspaceID() string {
	return e.WorkspaceID
}

func NewBaseEvent(eventType EventType, workspaceID common.ID) BaseEvent {
	return BaseEvent{
		ID:          common.NewID().String(),
		Type:        eventType,
		WorkspaceID: workspaceID.String(),
		Timestamp:   time.Now(),
	}
}

// ItemCreated event
type ItemCreated struct {
	BaseEvent
	ItemID    string
	VersionID string
	ProjectID string
}

func NewItemCreated(workspaceID, itemID, versionID, projectID common.ID) *ItemCreated {
	return &ItemCreated{
		BaseEvent: NewBaseEvent(ItemCreatedEvent, workspaceID),
		ItemID:    itemID.String(),
		VersionID: versionID.String(),
		ProjectID: projectID.String(),
	}
}

// VersionCreated event
type VersionCreated struct {
	BaseEvent
	ItemID    string
	VersionID string
	Version   int
}

func NewVersionCreated(workspaceID, itemID, versionID common.ID, version int) *VersionCreated {
	return &VersionCreated{
		BaseEvent: NewBaseEvent(VersionCreatedEvent, workspaceID),
		ItemID:    itemID.String(),
		VersionID: versionID.String(),
		Version:   version,
	}
}

// PreviousVersionDeprecated event
type PreviousVersionDeprecated struct {
	BaseEvent
	ItemID         string
	DeprecatedVersion int
}

func NewPreviousVersionDeprecated(workspaceID, itemID common.ID, deprecatedVersion int) *PreviousVersionDeprecated {
	return &PreviousVersionDeprecated{
		BaseEvent:      NewBaseEvent(PreviousVersionDeprecatedEvent, workspaceID),
		ItemID:         itemID.String(),
		DeprecatedVersion: deprecatedVersion,
	}
}

// ItemPublished event
type ItemPublished struct {
	BaseEvent
	ItemID    string
	VersionID string
}

func NewItemPublished(workspaceID, itemID, versionID common.ID) *ItemPublished {
	return &ItemPublished{
		BaseEvent: NewBaseEvent(ItemPublishedEvent, workspaceID),
		ItemID:    itemID.String(),
		VersionID: versionID.String(),
	}
}

// SourceLinkedToItem event
type SourceLinkedToItem struct {
	BaseEvent
	ItemID   string
	SourceID string
}

func NewSourceLinkedToItem(workspaceID, itemID, sourceID common.ID) *SourceLinkedToItem {
	return &SourceLinkedToItem{
		BaseEvent: NewBaseEvent(SourceLinkedToItemEvent, workspaceID),
		ItemID:    itemID.String(),
		SourceID:  sourceID.String(),
	}
}

// RelationCreated event
type RelationCreated struct {
	BaseEvent
	RelationID     string
	FromItemID     string
	ToItemID       string
	RelationTypeID string
}

func NewRelationCreated(workspaceID, relationID, fromItemID, toItemID, relationTypeID common.ID) *RelationCreated {
	return &RelationCreated{
		BaseEvent:      NewBaseEvent(RelationCreatedEvent, workspaceID),
		RelationID:     relationID.String(),
		FromItemID:     fromItemID.String(),
		ToItemID:       toItemID.String(),
		RelationTypeID: relationTypeID.String(),
	}
}

// EmbeddingRequested event
type EmbeddingRequested struct {
	BaseEvent
	ItemID    string
	VersionID string
	Content   string
}

func NewEmbeddingRequested(workspaceID, itemID, versionID common.ID, content string) *EmbeddingRequested {
	return &EmbeddingRequested{
		BaseEvent: NewBaseEvent(EmbeddingRequestedEvent, workspaceID),
		ItemID:    itemID.String(),
		VersionID: versionID.String(),
		Content:   content,
	}
}

// EmbeddingGenerated event
type EmbeddingGenerated struct {
	BaseEvent
	VersionID string
	Embedding []float32
}

func NewEmbeddingGenerated(workspaceID, versionID common.ID, embedding []float32) *EmbeddingGenerated {
	return &EmbeddingGenerated{
		BaseEvent: NewBaseEvent(EmbeddingGeneratedEvent, workspaceID),
		VersionID: versionID.String(),
		Embedding: embedding,
	}
}

// ChunksCreated event - fired when document chunks are created
type ChunksCreated struct {
	BaseEvent
	ItemID        string
	VersionID     string
	ChunkIDs      []string
	ChunkCount    int
}

func NewChunksCreated(workspaceID, itemID, versionID common.ID, chunkIDs []common.ID) *ChunksCreated {
	strIDs := make([]string, len(chunkIDs))
	for i, id := range chunkIDs {
		strIDs[i] = id.String()
	}

	return &ChunksCreated{
		BaseEvent:  NewBaseEvent(ChunksCreatedEvent, workspaceID),
		ItemID:     itemID.String(),
		VersionID:  versionID.String(),
		ChunkIDs:   strIDs,
		ChunkCount: len(chunkIDs),
	}
}

// ChunkEmbeddingsRequested event - fired when embeddings are requested for chunks
type ChunkEmbeddingsRequested struct {
	BaseEvent
	ItemID     string
	VersionID  string
	ChunkIDs   []string
	ChunkCount int
}

func NewChunkEmbeddingsRequested(workspaceID, itemID, versionID common.ID, chunkIDs []common.ID) *ChunkEmbeddingsRequested {
	strIDs := make([]string, len(chunkIDs))
	for i, id := range chunkIDs {
		strIDs[i] = id.String()
	}

	return &ChunkEmbeddingsRequested{
		BaseEvent:  NewBaseEvent(ChunkEmbeddingsRequestedEvent, workspaceID),
		ItemID:     itemID.String(),
		VersionID:  versionID.String(),
		ChunkIDs:   strIDs,
		ChunkCount: len(chunkIDs),
	}
}

// ChunkEmbeddingsGenerated event - fired when embeddings are generated for chunks
type ChunkEmbeddingsGenerated struct {
	BaseEvent
	VersionID     string
	ChunkCount    int
	TokensUsed    int
	SuccessCount  int
	FailureCount  int
}

func NewChunkEmbeddingsGenerated(workspaceID, versionID common.ID, chunkCount, tokensUsed, successCount, failureCount int) *ChunkEmbeddingsGenerated {
	return &ChunkEmbeddingsGenerated{
		BaseEvent:    NewBaseEvent(ChunkEmbeddingsGeneratedEvent, workspaceID),
		VersionID:    versionID.String(),
		ChunkCount:   chunkCount,
		TokensUsed:   tokensUsed,
		SuccessCount: successCount,
		FailureCount: failureCount,
	}
}
