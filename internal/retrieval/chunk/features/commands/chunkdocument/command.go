package chunkdocument

import (
	"database/sql"

	shared "documind.jordi.org/internal/shared/domain"
)

// Command holds input for chunking a document.
type Command struct {
	ItemID      shared.ID
	VersionID   shared.ID
	WorkspaceID shared.ID
	Title       string
	Summary     string
	BodyMd      string
}

// --- helpers ---

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
