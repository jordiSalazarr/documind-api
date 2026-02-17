package getchunk

import (
	"errors"

	shared "documind.jordi.org/internal/shared/domain"
)

// ErrChunkNotFound is returned when a chunk is not found.
var ErrChunkNotFound = errors.New("chunk not found")

// Query holds input for retrieving a single chunk.
type Query struct {
	ID          string
	IncludeItem bool
}

// Result is the response payload for a single chunk.
type Result struct {
	ID            string         `json:"id"`
	Content       string         `json:"content"`
	Heading       string         `json:"heading"`
	ChunkIndex    int            `json:"chunk_index"`
	TokenCount    int            `json:"token_count"`
	CharCount     int            `json:"char_count"`
	ItemVersionID string         `json:"item_version_id"`
	Metadata      MetadataResult `json:"metadata"`
	Item          *ItemResult    `json:"item,omitempty"`
}

// MetadataResult holds chunk metadata in the response.
type MetadataResult struct {
	StartCharOffset int    `json:"start_char_offset"`
	EndCharOffset   int    `json:"end_char_offset"`
	CodeBlock       bool   `json:"code_block"`
	Language        string `json:"language,omitempty"`
}

// ItemResult holds optional item info in the response.
type ItemResult struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	ProjectID string `json:"project_id"`
	Status    string `json:"status"`
	BodyMd    string `json:"body_md,omitempty"`
}

// ItemVersionReader is a cross-slice interface for getting item version info.
type ItemVersionReader interface {
	GetItemVersionTitle(itemVersionID shared.ID) (title, summary, projectID, status, bodyMd string, itemID shared.ID, err error)
}
