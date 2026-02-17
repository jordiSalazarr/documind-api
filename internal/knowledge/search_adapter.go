package knowledge

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"

	retrievaldomain "documind.jordi.org/internal/retrieval/domain"
	shareddomain "documind.jordi.org/internal/shared/domain"
)

// SearchAdapter implements the retrieval context's ItemSearchRepository using PostgreSQL.
type SearchAdapter struct {
	db *sql.DB
}

// NewSearchAdapter constructs a SearchAdapter.
func NewSearchAdapter(db *sql.DB) *SearchAdapter {
	return &SearchAdapter{db: db}
}

// SearchItemVersions performs full-text search on item versions within a workspace.
func (s *SearchAdapter) SearchItemVersions(ctx context.Context, workspaceID shareddomain.ID, query string, limit, offset int) ([]*retrievaldomain.ItemVersionSearchResult, error) {
	sqlQuery := `
		SELECT
			iv.id, iv.item_id, iv.workspace_id, iv.version,
			iv.title, iv.summary, iv.body_md, iv.custom_fields, iv.tags, iv.status,
			iv.created_at, iv.created_by,
			ts_rank(iv.search_vector, websearch_to_tsquery('english', $2)) AS rank
		FROM item_versions iv
		JOIN items i ON iv.item_id = i.id
		WHERE iv.workspace_id = $1
			AND i.deleted_at IS NULL
			AND iv.search_vector @@ websearch_to_tsquery('english', $2)
		ORDER BY rank DESC, iv.created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := s.db.QueryContext(ctx, sqlQuery, workspaceID, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search item versions: %w", err)
	}
	defer rows.Close()

	return scanSearchResults(rows)
}

// SearchItemVersionsByProject performs full-text search on item versions within a project.
func (s *SearchAdapter) SearchItemVersionsByProject(ctx context.Context, projectID shareddomain.ID, query string, limit, offset int) ([]*retrievaldomain.ItemVersionSearchResult, error) {
	sqlQuery := `
		SELECT
			iv.id, iv.item_id, iv.workspace_id, iv.version,
			iv.title, iv.summary, iv.body_md, iv.custom_fields, iv.tags, iv.status,
			iv.created_at, iv.created_by,
			ts_rank(iv.search_vector, websearch_to_tsquery('english', $2)) AS rank
		FROM item_versions iv
		JOIN items i ON iv.item_id = i.id
		WHERE i.project_id = $1
			AND i.deleted_at IS NULL
			AND iv.search_vector @@ websearch_to_tsquery('english', $2)
		ORDER BY rank DESC, iv.created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := s.db.QueryContext(ctx, sqlQuery, projectID, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search item versions by project: %w", err)
	}
	defer rows.Close()

	return scanSearchResults(rows)
}

// SearchByEmbedding performs vector similarity search on item versions within a workspace.
func (s *SearchAdapter) SearchByEmbedding(ctx context.Context, workspaceID shareddomain.ID, embedding []float32, limit int, minSimilarity float64) ([]*retrievaldomain.ItemVersionSearchResult, error) {
	embeddingStr := vectorToString(embedding)

	sqlQuery := `
		SELECT
			iv.id, iv.item_id, iv.workspace_id, iv.version,
			iv.title, iv.summary, iv.body_md, iv.custom_fields, iv.tags, iv.status,
			iv.created_at, iv.created_by,
			1 - (iv.embedding <=> $2::vector) AS similarity
		FROM item_versions iv
		JOIN items i ON iv.item_id = i.id
		WHERE iv.workspace_id = $1
			AND i.deleted_at IS NULL
			AND iv.embedding IS NOT NULL
			AND 1 - (iv.embedding <=> $2::vector) >= $3
		ORDER BY iv.embedding <=> $2::vector
		LIMIT $4
	`

	rows, err := s.db.QueryContext(ctx, sqlQuery, workspaceID, embeddingStr, minSimilarity, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search by embedding: %w", err)
	}
	defer rows.Close()

	return scanSearchResults(rows)
}

// SearchByEmbeddingInProject performs vector similarity search on item versions within a project.
func (s *SearchAdapter) SearchByEmbeddingInProject(ctx context.Context, projectID shareddomain.ID, embedding []float32, limit int, minSimilarity float64) ([]*retrievaldomain.ItemVersionSearchResult, error) {
	embeddingStr := vectorToString(embedding)

	sqlQuery := `
		SELECT
			iv.id, iv.item_id, iv.workspace_id, iv.version,
			iv.title, iv.summary, iv.body_md, iv.custom_fields, iv.tags, iv.status,
			iv.created_at, iv.created_by,
			1 - (iv.embedding <=> $2::vector) AS similarity
		FROM item_versions iv
		JOIN items i ON iv.item_id = i.id
		WHERE i.project_id = $1
			AND i.deleted_at IS NULL
			AND iv.embedding IS NOT NULL
			AND 1 - (iv.embedding <=> $2::vector) >= $3
		ORDER BY iv.embedding <=> $2::vector
		LIMIT $4
	`

	rows, err := s.db.QueryContext(ctx, sqlQuery, projectID, embeddingStr, minSimilarity, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search by embedding in project: %w", err)
	}
	defer rows.Close()

	return scanSearchResults(rows)
}

func scanSearchResults(rows *sql.Rows) ([]*retrievaldomain.ItemVersionSearchResult, error) {
	var results []*retrievaldomain.ItemVersionSearchResult
	for rows.Next() {
		var result retrievaldomain.ItemVersionSearchResult
		var customFieldsJSON []byte
		var createdAt sql.NullTime
		var score float64

		err := rows.Scan(
			&result.ID, &result.ItemID, &result.WorkspaceID, &result.Version,
			&result.Title, &result.Summary, &result.BodyMd, &customFieldsJSON,
			pq.Array(&result.Tags), &result.Status,
			&createdAt, &result.CreatedBy,
			&score,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

		if createdAt.Valid {
			result.CreatedAt = createdAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}

		if len(customFieldsJSON) > 0 {
			if err := json.Unmarshal(customFieldsJSON, &result.CustomFields); err != nil {
				return nil, fmt.Errorf("failed to unmarshal custom fields: %w", err)
			}
		}

		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search results: %w", err)
	}

	return results, nil
}

// vectorToString converts a float32 slice to PostgreSQL vector format.
func vectorToString(v []float32) string {
	s := make([]string, len(v))
	for i, f := range v {
		s[i] = fmt.Sprintf("%f", f)
	}
	return "[" + strings.Join(s, ",") + "]"
}
