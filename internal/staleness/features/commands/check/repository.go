package check

import (
	"context"
	"database/sql"

	"documind.jordi.org/internal/staleness/domain"
)

// Repository defines persistence for staleness check results.
type Repository interface {
	InsertReport(ctx context.Context, report *domain.StalenessReport) error
	InsertReportItems(ctx context.Context, items []*domain.StalenessReportItem) error
}

type postgresRepo struct{ db *sql.DB }

func newPostgresRepo(db *sql.DB) *postgresRepo { return &postgresRepo{db: db} }

func (r *postgresRepo) InsertReport(ctx context.Context, report *domain.StalenessReport) error {
	query := `
		INSERT INTO staleness_reports (id, workspace_id, project_id, trigger_type, trigger_ref, repository_url, diff_summary, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	var projectID *string
	if report.ProjectID != nil {
		s := report.ProjectID.String()
		projectID = &s
	}

	_, err := r.db.ExecContext(ctx, query,
		report.ID.String(),
		report.WorkspaceID.String(),
		projectID,
		string(report.TriggerType),
		report.TriggerRef,
		report.RepositoryURL,
		report.DiffSummary,
		string(report.Status),
		report.CreatedAt,
		report.UpdatedAt,
	)
	return err
}

func (r *postgresRepo) InsertReportItems(ctx context.Context, items []*domain.StalenessReportItem) error {
	if len(items) == 0 {
		return nil
	}

	query := `
		INSERT INTO staleness_report_items (id, report_id, item_id, item_version_id, item_title, confidence, explanation, suggested_update, stale, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	for _, item := range items {
		_, err := r.db.ExecContext(ctx, query,
			item.ID.String(),
			item.ReportID.String(),
			item.ItemID.String(),
			item.ItemVersionID.String(),
			item.ItemTitle,
			item.Confidence,
			item.Explanation,
			item.SuggestedUpdate,
			item.Stale,
			item.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
