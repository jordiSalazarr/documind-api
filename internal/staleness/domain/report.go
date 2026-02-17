package domain

import (
	"time"

	shared "documind.jordi.org/internal/shared/domain"
)

// StalenessStatus represents the lifecycle status of a staleness report.
type StalenessStatus string

const (
	StatusOpen      StalenessStatus = "open"
	StatusDismissed StalenessStatus = "dismissed"
	StatusResolved  StalenessStatus = "resolved"
)

// TriggerType represents what triggered a staleness check.
type TriggerType string

const (
	TriggerCodeChange TriggerType = "code_change"
	TriggerManual     TriggerType = "manual"
)

// StalenessReport represents a staleness detection run.
type StalenessReport struct {
	ID            shared.ID
	WorkspaceID   shared.ID
	ProjectID     *shared.ID
	TriggerType   TriggerType
	TriggerRef    string
	RepositoryURL string
	DiffSummary   string
	Status        StalenessStatus
	Items         []*StalenessReportItem
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// StalenessReportItem represents a single document evaluated for staleness.
type StalenessReportItem struct {
	ID              shared.ID
	ReportID        shared.ID
	ItemID          shared.ID
	ItemVersionID   shared.ID
	ItemTitle       string
	Confidence      float64
	Explanation     string
	SuggestedUpdate string
	Stale           bool
	CreatedAt       time.Time
}

// NewReport creates a new staleness report.
func NewReport(workspaceID shared.ID, projectID *shared.ID, triggerType TriggerType, triggerRef, repoURL, diffSummary string) *StalenessReport {
	now := time.Now().UTC()
	return &StalenessReport{
		ID:            shared.NewID(),
		WorkspaceID:   workspaceID,
		ProjectID:     projectID,
		TriggerType:   triggerType,
		TriggerRef:    triggerRef,
		RepositoryURL: repoURL,
		DiffSummary:   diffSummary,
		Status:        StatusOpen,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// NewReportItem creates a new staleness report item.
func NewReportItem(reportID, itemID, itemVersionID shared.ID, itemTitle string, stale bool, confidence float64, explanation, suggestedUpdate string) *StalenessReportItem {
	return &StalenessReportItem{
		ID:              shared.NewID(),
		ReportID:        reportID,
		ItemID:          itemID,
		ItemVersionID:   itemVersionID,
		ItemTitle:       itemTitle,
		Confidence:      confidence,
		Explanation:     explanation,
		SuggestedUpdate: suggestedUpdate,
		Stale:           stale,
		CreatedAt:       time.Now().UTC(),
	}
}
