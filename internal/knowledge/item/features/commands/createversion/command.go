package createversion

import (
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// Command holds the data needed to create a new version of an item.
type Command struct {
	ItemID       shareddomain.ID
	WorkspaceID  shareddomain.ID
	Title        string
	Summary      string
	BodyMd       string
	CustomFields map[string]interface{}
	Tags         []string
	CreatedBy    shareddomain.ID
}

// --- Cross-slice interfaces ---

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
