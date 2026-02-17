package list

import (
	"context"
	"database/sql"
	"fmt"

	projdomain "documind.jordi.org/internal/knowledge/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

type Repository interface {
	ListByWorkspace(ctx context.Context, workspaceID shared.ID, limit, offset int) ([]*projdomain.Project, error)
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) ListByWorkspace(ctx context.Context, workspaceID shared.ID, limit, offset int) ([]*projdomain.Project, error) {
	query := `
		SELECT id, workspace_id, name, slug, description, created_at, updated_at, created_by, updated_by
		FROM projects WHERE workspace_id = $1 ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, workspaceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()
	var projects []*projdomain.Project
	for rows.Next() {
		var p projdomain.Project
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Slug, &p.Description, &p.CreatedAt, &p.UpdatedAt, &p.CreatedBy, &p.UpdatedBy); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, &p)
	}
	return projects, rows.Err()
}
