package getreport

import (
	"context"
	"database/sql"
	"errors"
)

// Repository defines read access for a single staleness report.
type Repository interface {
	GetByID(ctx context.Context, id, workspaceID string) (*Result, error)
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) GetByID(ctx context.Context, id, workspaceID string) (*Result, error) {
	// Fetch report
	reportQuery := `
		SELECT id, project_id, trigger_type, trigger_ref, repository_url, diff_summary, status, created_at, updated_at
		FROM staleness_reports
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`
	var result Result
	var projectID sql.NullString
	err := r.db.QueryRowContext(ctx, reportQuery, id, workspaceID).Scan(
		&result.ID, &projectID, &result.TriggerType, &result.TriggerRef,
		&result.RepositoryURL, &result.DiffSummary, &result.Status,
		&result.CreatedAt, &result.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrReportNotFound
	}
	if err != nil {
		return nil, err
	}
	if projectID.Valid {
		result.ProjectID = &projectID.String
	}

	// Fetch items sorted by stale DESC, confidence DESC
	itemsQuery := `
		SELECT id, item_id, item_version_id, item_title, stale, confidence, explanation, suggested_update, created_at
		FROM staleness_report_items
		WHERE report_id = $1
		ORDER BY stale DESC, confidence DESC
	`
	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item ItemEntry
		var suggestedUpdate sql.NullString
		if err := rows.Scan(
			&item.ID, &item.ItemID, &item.ItemVersionID, &item.ItemTitle,
			&item.Stale, &item.Confidence, &item.Explanation, &suggestedUpdate,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if suggestedUpdate.Valid {
			item.SuggestedUpdate = suggestedUpdate.String
		}
		result.Items = append(result.Items, &item)
	}
	if result.Items == nil {
		result.Items = []*ItemEntry{}
	}

	return &result, rows.Err()
}
