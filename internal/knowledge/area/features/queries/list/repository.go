package list

import (
	"context"
	"database/sql"
	"fmt"

	svcdomain "documind.jordi.org/internal/knowledge/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

type Repository interface {
	ListByProject(ctx context.Context, projectID shared.ID, limit, offset int) ([]*svcdomain.Area, error)
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) ListByProject(ctx context.Context, projectID shared.ID, limit, offset int) ([]*svcdomain.Area, error) {
	query := `
		SELECT id, project_id, name, slug, description, created_at, updated_at
		FROM services WHERE project_id = $1 ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list areas: %w", err)
	}
	defer rows.Close()
	var areas []*svcdomain.Area
	for rows.Next() {
		var area svcdomain.Area
		if err := rows.Scan(&area.ID, &area.ProjectID, &area.Name, &area.Slug, &area.Description, &area.CreatedAt, &area.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan area: %w", err)
		}
		areas = append(areas, &area)
	}
	return areas, rows.Err()
}
