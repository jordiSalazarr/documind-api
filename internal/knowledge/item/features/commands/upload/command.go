package upload

import (
	"context"
	"io"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
	"documind.jordi.org/internal/knowledge/parser"
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// Command holds the data needed to upload a file as a new item.
type Command struct {
	WorkspaceID shareddomain.ID
	ProjectID   shareddomain.ID
	AreaID      *shareddomain.ID // optional
	ItemTypeID  shareddomain.ID
	Filename    string
	FileSize    int64
	FileReader  io.Reader
	OwnerUserID shareddomain.ID
	CreatedBy   shareddomain.ID
}

// Result holds the output of a successful file upload.
type Result struct {
	Item    *itemdomain.Item
	Version *itemdomain.ItemVersion
	Parsing ParsingResult
}

// ParsingResult contains metadata about how the file was parsed.
type ParsingResult struct {
	SourceFormat string `json:"source_format"`
	PageCount    int    `json:"page_count,omitempty"`
	WordCount    int    `json:"word_count"`
	Status       string `json:"status"` // "chunking_in_progress"
}

// --- Cross-context interfaces ---

// DocumentChunker triggers document chunking for an item version.
type DocumentChunker interface {
	ChunkDocument(
		itemID shareddomain.ID,
		versionID shareddomain.ID,
		workspaceID shareddomain.ID,
		title string,
		summary string,
		bodyMd string,
	) error
}

// FileParser parses uploaded files into structured documents.
type FileParser interface {
	Parse(ctx context.Context, filename string, reader io.Reader) (*parser.ParsedDocument, error)
}

// FileValidator validates uploaded files.
type FileValidator interface {
	Validate(filename string, size int64) error
}
