package staleness

import (
	"context"
	"database/sql"
	"fmt"

	"documind.jordi.org/internal/retrieval/search/features/queries/retrieve"
	"documind.jordi.org/internal/staleness/features/commands/check"

	shared "documind.jordi.org/internal/shared/domain"
)

// RetrieverAdapter bridges retrieve.Handler to check.Retriever interface,
// enriching chunks with item_id and title from item_versions.
type RetrieverAdapter struct {
	handler *retrieve.Handler
	db      *sql.DB
}

// NewRetrieverAdapter creates a new retriever adapter.
func NewRetrieverAdapter(handler *retrieve.Handler, db *sql.DB) *RetrieverAdapter {
	return &RetrieverAdapter{handler: handler, db: db}
}

// Retrieve calls the hybrid retrieval handler and enriches results with item info.
func (a *RetrieverAdapter) Retrieve(ctx context.Context, input check.RetrievalInput) ([]*check.RetrievalResult, error) {
	// Build retrieve input
	rInput := retrieve.Input{
		Query:       input.Query,
		WorkspaceID: shared.ID(input.WorkspaceID),
		MaxResults:  input.MaxResults,
	}
	if input.ProjectID != "" {
		pid := shared.ID(input.ProjectID)
		rInput.ProjectID = &pid
	}

	output, err := a.handler.Handle(ctx, rInput)
	if err != nil {
		return nil, fmt.Errorf("retrieval failed: %w", err)
	}

	if len(output.Results) == 0 {
		return nil, nil
	}

	// Collect item_version_ids for batch enrichment
	versionIDs := make([]string, len(output.Results))
	for i, r := range output.Results {
		versionIDs[i] = r.Chunk.ItemVersionID.String()
	}

	// Batch query item_versions for item_id and title
	enrichment, err := a.enrichItemVersions(ctx, versionIDs)
	if err != nil {
		// Log but continue with un-enriched results
		enrichment = map[string]itemInfo{}
	}

	results := make([]*check.RetrievalResult, len(output.Results))
	for i, r := range output.Results {
		versionID := r.Chunk.ItemVersionID.String()
		info := enrichment[versionID]
		results[i] = &check.RetrievalResult{
			ChunkID:       r.Chunk.ID.String(),
			ItemID:        info.ItemID,
			ItemVersionID: versionID,
			ItemTitle:     info.Title,
			Content:       r.Chunk.Content,
			FinalScore:    r.FinalScore,
		}
	}

	return results, nil
}

type itemInfo struct {
	ItemID string
	Title  string
}

func (a *RetrieverAdapter) enrichItemVersions(ctx context.Context, versionIDs []string) (map[string]itemInfo, error) {
	if len(versionIDs) == 0 {
		return nil, nil
	}

	query := `SELECT iv.id, iv.item_id, iv.title FROM item_versions iv WHERE iv.id = ANY($1::uuid[])`

	// Convert to pq array format
	rows, err := a.db.QueryContext(ctx, query, pqStringArray(versionIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]itemInfo)
	for rows.Next() {
		var id, itemID, title string
		if err := rows.Scan(&id, &itemID, &title); err != nil {
			return nil, err
		}
		result[id] = itemInfo{ItemID: itemID, Title: title}
	}
	return result, rows.Err()
}

// pqStringArray formats a string slice as a PostgreSQL array literal.
func pqStringArray(ss []string) string {
	if len(ss) == 0 {
		return "{}"
	}
	return "{" + joinStrings(ss) + "}"
}

func joinStrings(ss []string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += ","
		}
		result += s
	}
	return result
}
