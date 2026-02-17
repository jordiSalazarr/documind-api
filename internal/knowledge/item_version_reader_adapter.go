package knowledge

import (
	"context"
	"database/sql"
	"fmt"

	shareddomain "documind.jordi.org/internal/shared/domain"
)

// ItemVersionReaderAdapter satisfies the retrieval context's ItemVersionReader interface
// by querying item versions directly.
type ItemVersionReaderAdapter struct {
	db *sql.DB
}

// NewItemVersionReaderAdapter creates a new adapter.
func NewItemVersionReaderAdapter(db *sql.DB) *ItemVersionReaderAdapter {
	return &ItemVersionReaderAdapter{db: db}
}

// GetItemVersionTitle returns title, summary, projectID, status, bodyMd, and itemID
// for a given item version.
func (a *ItemVersionReaderAdapter) GetItemVersionTitle(itemVersionID shareddomain.ID) (title, summary, projectID, status, bodyMd string, itemID shareddomain.ID, err error) {
	query := `
		SELECT iv.title, iv.summary, i.project_id, iv.status, iv.body_md, iv.item_id
		FROM item_versions iv
		JOIN items i ON iv.item_id = i.id
		WHERE iv.id = $1
	`

	err = a.db.QueryRowContext(context.Background(), query, itemVersionID).Scan(
		&title, &summary, &projectID, &status, &bodyMd, &itemID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", "", "", "", "", fmt.Errorf("item version not found: %s", itemVersionID)
		}
		return "", "", "", "", "", "", fmt.Errorf("failed to get item version title: %w", err)
	}

	return title, summary, projectID, status, bodyMd, itemID, nil
}
