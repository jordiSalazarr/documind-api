package updatereport

import "errors"

var (
	ErrReportNotFound = errors.New("staleness report not found")
	ErrInvalidStatus  = errors.New("status must be 'dismissed' or 'resolved'")
)

// Command contains parameters for updating a report's status.
type Command struct {
	ID          string
	WorkspaceID string
	Status      string // "dismissed" or "resolved"
}
