package retrieval

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pgvector/pgvector-go"

	shared "documind.jordi.org/internal/shared/domain"
	"documind.jordi.org/internal/shared/workers"
)

// WorkerAdapter adapts the chunk DB layer to satisfy the
// workers.ChunkEmbeddingRepository interface.
type WorkerAdapter struct {
	db *sql.DB
}

// NewWorkerAdapter creates a new adapter.
func NewWorkerAdapter(db *sql.DB) *WorkerAdapter {
	return &WorkerAdapter{db: db}
}

// ListByItemVersionID returns chunks mapped to workers.ChunkForEmbedding.
func (a *WorkerAdapter) ListByItemVersionID(versionID shared.ID) ([]*workers.ChunkForEmbedding, error) {
	ctx := context.Background()
	query := `
		SELECT id, item_version_id, workspace_id, content,
			chunk_index, chunk_level, token_count, char_count,
			heading, parent_chunk_id, embedding, metadata, created_at, deleted_at
		FROM chunks
		WHERE item_version_id = $1 AND deleted_at IS NULL
		ORDER BY chunk_index ASC
	`

	rows, err := a.db.QueryContext(ctx, query, versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list chunks: %w", err)
	}
	defer rows.Close()

	var result []*workers.ChunkForEmbedding
	for rows.Next() {
		var (
			id             shared.ID
			itemVersionID  shared.ID
			workspaceID    shared.ID
			content        string
			chunkIndex     int
			chunkLevel     int
			tokenCount     int
			charCount      int
			heading        sql.NullString
			parentChunkID  *shared.ID
			embeddingBytes []byte
			metadataJSON   []byte
			createdAt      sql.NullTime
			deletedAt      sql.NullTime
		)

		err := rows.Scan(
			&id, &itemVersionID, &workspaceID, &content,
			&chunkIndex, &chunkLevel, &tokenCount, &charCount,
			&heading, &parentChunkID, &embeddingBytes, &metadataJSON,
			&createdAt, &deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chunk: %w", err)
		}

		var embedding []float32
		if embeddingBytes != nil {
			var vec pgvector.Vector
			if err := vec.Scan(embeddingBytes); err == nil {
				embedding = vec.Slice()
			}
		}

		result = append(result, &workers.ChunkForEmbedding{
			ID:        id,
			Content:   content,
			Embedding: embedding,
		})
	}

	return result, nil
}

// UpdateEmbeddingsBatch batch-updates embeddings for chunks.
func (a *WorkerAdapter) UpdateEmbeddingsBatch(updates map[shared.ID][]float32) error {
	if len(updates) == 0 {
		return nil
	}

	ctx := context.Background()
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `UPDATE chunks SET embedding = $1 WHERE id = $2`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for chunkID, embedding := range updates {
		vec := pgvector.NewVector(embedding)
		_, err := stmt.ExecContext(ctx, vec, chunkID)
		if err != nil {
			return fmt.Errorf("failed to update embedding for chunk %s: %w", chunkID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
