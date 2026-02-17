package listreports

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository defines read access for staleness reports.
type Repository interface {
	List(ctx context.Context, q Query) ([]*Result, error)
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) List(ctx context.Context, q Query) ([]*Result, error) {
	baseQuery := `
		SELECT
			sr.id, sr.project_id, sr.trigger_type, sr.trigger_ref,
			sr.repository_url, sr.diff_summary, sr.status,
			COALESCE(SUM(CASE WHEN sri.stale THEN 1 ELSE 0 END), 0)::int AS stale_count,
			COUNT(sri.id)::int AS total_checked,
			sr.created_at, sr.updated_at
		FROM staleness_reports sr
		LEFT JOIN staleness_report_items sri ON sri.report_id = sr.id
		WHERE sr.workspace_id = $1 AND sr.deleted_at IS NULL
	`
	args := []interface{}{q.WorkspaceID}
	argIdx := 2

	if q.ProjectID != "" {
		baseQuery += fmt.Sprintf(" AND sr.project_id = $%d", argIdx)
		args = append(args, q.ProjectID)
		argIdx++
	}
	if q.Status != "" {
		baseQuery += fmt.Sprintf(" AND sr.status = $%d", argIdx)
		args = append(args, q.Status)
		argIdx++
	}

	baseQuery += `
		GROUP BY sr.id
		ORDER BY sr.created_at DESC
	`

	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, q.Limit, (q.Page-1)*q.Limit)

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*Result
	for rows.Next() {
		var res Result
		var projectID sql.NullString
		if err := rows.Scan(
			&res.ID, &projectID, &res.TriggerType, &res.TriggerRef,
			&res.RepositoryURL, &res.DiffSummary, &res.Status,
			&res.StaleCount, &res.TotalChecked,
			&res.CreatedAt, &res.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if projectID.Valid {
			res.ProjectID = &projectID.String
		}
		results = append(results, &res)
	}
	return results, rows.Err()
}
