package create

import (
	itemdomain "documind.jordi.org/internal/knowledge/domain"
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// --- Command & Result ---

// Command holds the data needed to create an item with its first version.
type Command struct {
	WorkspaceID  shareddomain.ID
	ProjectID    shareddomain.ID
	ServiceID    *shareddomain.ID
	ItemTypeID   shareddomain.ID
	Title        string
	Summary      string
	BodyMd       string
	CustomFields map[string]interface{}
	Tags         []string
	OwnerUserID  shareddomain.ID
	CreatedBy    shareddomain.ID
}

// Result holds the output of a successful item creation.
type Result struct {
	Item    *itemdomain.Item
	Version *itemdomain.ItemVersion
}

// --- Cross-context interface ---

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
