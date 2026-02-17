package domain

import (
	"time"

	shared "documind.jordi.org/internal/shared/domain"
)

// ChunkLevel represents the hierarchical level of a chunk
type ChunkLevel int

const (
	ChunkLevelSection   ChunkLevel = 1 // Section-based chunks (largest)
	ChunkLevelParagraph ChunkLevel = 2 // Paragraph-based chunks (medium)
	ChunkLevelSentence  ChunkLevel = 3 // Sentence-based chunks (smallest)
)

// Chunk represents a document chunk for RAG retrieval
type Chunk struct {
	ID            shared.ID
	ItemVersionID shared.ID
	WorkspaceID   shared.ID
	Content       string
	ChunkIndex    int
	ChunkLevel    ChunkLevel
	TokenCount    int
	CharCount     int
	Heading       string     // Section title if applicable
	ParentChunkID *shared.ID // For hierarchical chunks
	Embedding     []float32  // 1536 dimensions for text-embedding-3-small
	Metadata      ChunkMetadata
	CreatedAt     time.Time
	DeletedAt     *time.Time
}

// ChunkMetadata stores additional chunk information
type ChunkMetadata struct {
	StartCharOffset int     `json:"start_char_offset"`
	EndCharOffset   int     `json:"end_char_offset"`
	CodeBlock       bool    `json:"code_block"`
	Language        string  `json:"language,omitempty"`
	PageNumbers     []int   `json:"page_numbers,omitempty"`
	TableRef        *string `json:"table_ref,omitempty"`
}

// NewChunk creates a new chunk
func NewChunk(
	itemVersionID shared.ID,
	workspaceID shared.ID,
	content string,
	chunkIndex int,
	chunkLevel ChunkLevel,
	tokenCount int,
) *Chunk {
	return &Chunk{
		ID:            shared.NewID(),
		ItemVersionID: itemVersionID,
		WorkspaceID:   workspaceID,
		Content:       content,
		ChunkIndex:    chunkIndex,
		ChunkLevel:    chunkLevel,
		TokenCount:    tokenCount,
		CharCount:     len(content),
		Metadata:      ChunkMetadata{},
		CreatedAt:     time.Now(),
	}
}

// SetHeading sets the section heading for this chunk
func (c *Chunk) SetHeading(heading string) {
	c.Heading = heading
}

// SetParent sets the parent chunk for hierarchical chunking
func (c *Chunk) SetParent(parentID shared.ID) {
	c.ParentChunkID = &parentID
}

// SetEmbedding sets the embedding vector
func (c *Chunk) SetEmbedding(embedding []float32) {
	c.Embedding = embedding
}

// SetCharOffsets sets the start and end character offsets in the original document
func (c *Chunk) SetCharOffsets(start, end int) {
	c.Metadata.StartCharOffset = start
	c.Metadata.EndCharOffset = end
}

// MarkAsCodeBlock marks this chunk as containing code
func (c *Chunk) MarkAsCodeBlock(language string) {
	c.Metadata.CodeBlock = true
	c.Metadata.Language = language
}

// SetPageNumbers sets the page numbers this chunk appears on (for PDF documents)
func (c *Chunk) SetPageNumbers(pages []int) {
	c.Metadata.PageNumbers = pages
}

// SoftDelete marks the chunk as deleted
func (c *Chunk) SoftDelete() {
	now := time.Now()
	c.DeletedAt = &now
}

// IsDeleted checks if the chunk is soft-deleted
func (c *Chunk) IsDeleted() bool {
	return c.DeletedAt != nil
}
