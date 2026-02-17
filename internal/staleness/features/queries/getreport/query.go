package getreport

import "errors"

var (
	ErrReportNotFound = errors.New("staleness report not found")
)

// Query contains parameters for getting a staleness report.
type Query struct {
	ID          string
	WorkspaceID string
}

// Result contains the full report with all items.
type Result struct {
	ID            string       `json:"id"`
	ProjectID     *string      `json:"project_id,omitempty"`
	TriggerType   string       `json:"trigger_type"`
	TriggerRef    string       `json:"trigger_ref"`
	RepositoryURL string       `json:"repository_url"`
	DiffSummary   string       `json:"diff_summary"`
	Status        string       `json:"status"`
	Items         []*ItemEntry `json:"items"`
	CreatedAt     string       `json:"created_at"`
	UpdatedAt     string       `json:"updated_at"`
}

// ItemEntry represents a single evaluated item in the report.
type ItemEntry struct {
	ID              string  `json:"id"`
	ItemID          string  `json:"item_id"`
	ItemVersionID   string  `json:"item_version_id"`
	ItemTitle       string  `json:"item_title"`
	Stale           bool    `json:"stale"`
	Confidence      float64 `json:"confidence"`
	Explanation     string  `json:"explanation"`
	SuggestedUpdate string  `json:"suggested_update,omitempty"`
	CreatedAt       string  `json:"created_at"`
}
