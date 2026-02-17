package chunkdocument

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/pgvector/pgvector-go"

	chunkdomain "documind.jordi.org/internal/retrieval/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

// WriteRepository defines write operations needed by this use case.
type WriteRepository interface {
	DeleteByItemVersionID(versionID shared.ID) error
	CreateBatch(chunks []*chunkdomain.Chunk) error
}

// --- private postgres implementation ---

type postgresWriteRepo struct {
	db *sql.DB
}

func newPostgresWriteRepo(db *sql.DB) *postgresWriteRepo {
	return &postgresWriteRepo{db: db}
}

func (r *postgresWriteRepo) DeleteByItemVersionID(itemVersionID shared.ID) error {
	ctx := context.Background()
	query := `UPDATE chunks SET deleted_at = NOW() WHERE item_version_id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, itemVersionID)
	if err != nil {
		return fmt.Errorf("failed to delete chunks: %w", err)
	}

	return nil
}

func (r *postgresWriteRepo) CreateBatch(chunks []*chunkdomain.Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO chunks (
			id, item_version_id, workspace_id, content,
			chunk_index, chunk_level, token_count, char_count,
			heading, parent_chunk_id, embedding, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, chunk := range chunks {
		metadataJSON, err := json.Marshal(chunk.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		var embedding interface{}
		if chunk.Embedding != nil {
			embedding = pgvector.NewVector(chunk.Embedding)
		}

		_, err = stmt.ExecContext(ctx,
			chunk.ID, chunk.ItemVersionID, chunk.WorkspaceID, chunk.Content,
			chunk.ChunkIndex, chunk.ChunkLevel, chunk.TokenCount, chunk.CharCount,
			nullString(chunk.Heading), chunk.ParentChunkID, embedding,
			metadataJSON, chunk.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert chunk %d: %w", chunk.ChunkIndex, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
