package listreports

// Query contains parameters for listing staleness reports.
type Query struct {
	WorkspaceID string
	ProjectID   string // optional filter
	Status      string // optional filter: open, dismissed, resolved
	Page        int
	Limit       int
}

// Result represents a single report in the list.
type Result struct {
	ID            string  `json:"id"`
	ProjectID     *string `json:"project_id,omitempty"`
	TriggerType   string  `json:"trigger_type"`
	TriggerRef    string  `json:"trigger_ref"`
	RepositoryURL string  `json:"repository_url"`
	DiffSummary   string  `json:"diff_summary"`
	Status        string  `json:"status"`
	StaleCount    int     `json:"stale_count"`
	TotalChecked  int     `json:"total_checked"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}
