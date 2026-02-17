package retrieval

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/pgvector/pgvector-go"

	retrievaldomain "documind.jordi.org/internal/retrieval/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

// ChunkSearchAdapter satisfies retrieve.ChunkSearchRepository by providing
// SearchByEmbedding and SearchByText directly against the chunks table.
type ChunkSearchAdapter struct {
	db *sql.DB
}

// NewChunkSearchAdapter creates a new search adapter.
func NewChunkSearchAdapter(db *sql.DB) *ChunkSearchAdapter {
	return &ChunkSearchAdapter{db: db}
}

// SearchByEmbedding performs vector similarity search on chunks.
func (a *ChunkSearchAdapter) SearchByEmbedding(
	workspaceID shared.ID,
	embedding []float32,
	limit int,
	minSimilarity float64,
) ([]*retrievaldomain.Chunk, error) {
	ctx := context.Background()
	maxDistance := 1.0 - minSimilarity

	query := `
		SELECT id, item_version_id, workspace_id, content,
			chunk_index, chunk_level, token_count, char_count,
			heading, parent_chunk_id, embedding, metadata, created_at, deleted_at,
			1 - (embedding <=> $1) AS similarity
		FROM chunks
		WHERE workspace_id = $2
			AND deleted_at IS NULL
			AND embedding IS NOT NULL
			AND embedding <=> $1 < $3
		ORDER BY embedding <=> $1
		LIMIT $4
	`

	vec := pgvector.NewVector(embedding)
	rows, err := a.db.QueryContext(ctx, query, vec, workspaceID, maxDistance, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search by embedding: %w", err)
	}
	defer rows.Close()

	chunks := []*retrievaldomain.Chunk{}
	for rows.Next() {
		chunk := &retrievaldomain.Chunk{}
		var metadataJSON []byte
		var embeddingVec pgvector.Vector
		var heading sql.NullString
		var parentChunkID *shared.ID
		var similarity float64

		err := rows.Scan(
			&chunk.ID, &chunk.ItemVersionID, &chunk.WorkspaceID, &chunk.Content,
			&chunk.ChunkIndex, &chunk.ChunkLevel, &chunk.TokenCount, &chunk.CharCount,
			&heading, &parentChunkID, &embeddingVec, &metadataJSON,
			&chunk.CreatedAt, &chunk.DeletedAt, &similarity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chunk: %w", err)
		}

		if heading.Valid {
			chunk.Heading = heading.String
		}
		chunk.ParentChunkID = parentChunkID

		if len(embeddingVec.Slice()) > 0 {
			chunk.Embedding = embeddingVec.Slice()
		}

		if err := json.Unmarshal(metadataJSON, &chunk.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// SearchByText performs full-text search on chunks.
func (a *ChunkSearchAdapter) SearchByText(workspaceID shared.ID, query string, limit int) ([]*retrievaldomain.Chunk, error) {
	ctx := context.Background()
	sqlQuery := `
		SELECT id, item_version_id, workspace_id, content,
			chunk_index, chunk_level, token_count, char_count,
			heading, parent_chunk_id, embedding, metadata, created_at, deleted_at,
			ts_rank(to_tsvector('english', content), websearch_to_tsquery('english', $1)) AS rank
		FROM chunks
		WHERE workspace_id = $2
			AND deleted_at IS NULL
			AND to_tsvector('english', content) @@ websearch_to_tsquery('english', $1)
		ORDER BY rank DESC
		LIMIT $3
	`

	rows, err := a.db.QueryContext(ctx, sqlQuery, query, workspaceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search by text: %w", err)
	}
	defer rows.Close()

	chunks := []*retrievaldomain.Chunk{}
	for rows.Next() {
		chunk := &retrievaldomain.Chunk{}
		var metadataJSON []byte
		var embeddingBytes []byte
		var heading sql.NullString
		var parentChunkID *shared.ID
		var rank float64

		err := rows.Scan(
			&chunk.ID, &chunk.ItemVersionID, &chunk.WorkspaceID, &chunk.Content,
			&chunk.ChunkIndex, &chunk.ChunkLevel, &chunk.TokenCount, &chunk.CharCount,
			&heading, &parentChunkID, &embeddingBytes, &metadataJSON,
			&chunk.CreatedAt, &chunk.DeletedAt, &rank,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chunk: %w", err)
		}

		if heading.Valid {
			chunk.Heading = heading.String
		}
		chunk.ParentChunkID = parentChunkID

		if embeddingBytes != nil {
			var embedding pgvector.Vector
			if err := embedding.Scan(embeddingBytes); err == nil {
				chunk.Embedding = embedding.Slice()
			}
		}

		if err := json.Unmarshal(metadataJSON, &chunk.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
