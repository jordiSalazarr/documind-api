package retrieval

import (
	"database/sql"

	"documind.jordi.org/internal/retrieval/chunk/features/commands/chunkdocument"
	"documind.jordi.org/internal/retrieval/chunker"
	shared "documind.jordi.org/internal/shared/domain"
	"documind.jordi.org/internal/shared/infrastructure/eventbus"
)

// ChunkerAdapter adapts chunkdocument.Handler to satisfy
// the knowledge context's DocumentChunker interface.
type ChunkerAdapter struct {
	handler *chunkdocument.Handler
}

// NewChunkerAdapter creates a new adapter, constructing the
// chunkdocument.Handler internally.
func NewChunkerAdapter(db *sql.DB, c *chunker.SemanticChunker, bus eventbus.EventBus) *ChunkerAdapter {
	return &ChunkerAdapter{
		handler: chunkdocument.NewHandler(db, c, bus),
	}
}

// ChunkDocument satisfies knowledge context's DocumentChunker interface.
func (a *ChunkerAdapter) ChunkDocument(
	itemID shared.ID,
	versionID shared.ID,
	workspaceID shared.ID,
	title string,
	summary string,
	bodyMd string,
) error {
	return a.handler.Handle(chunkdocument.Command{
		ItemID:      itemID,
		VersionID:   versionID,
		WorkspaceID: workspaceID,
		Title:       title,
		Summary:     summary,
		BodyMd:      bodyMd,
	})
}
