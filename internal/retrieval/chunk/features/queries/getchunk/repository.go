package getchunk

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/pgvector/pgvector-go"

	chunkdomain "documind.jordi.org/internal/retrieval/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

// ReadRepository defines the read operations needed by this use case.
type ReadRepository interface {
	GetByID(id shared.ID) (*chunkdomain.Chunk, error)
}

// --- private postgres implementation ---

type postgresReadRepo struct {
	db *sql.DB
}

func newPostgresReadRepo(db *sql.DB) *postgresReadRepo {
	return &postgresReadRepo{db: db}
}

func (r *postgresReadRepo) GetByID(id shared.ID) (*chunkdomain.Chunk, error) {
	ctx := context.Background()
	query := `
		SELECT id, item_version_id, workspace_id, content,
			chunk_index, chunk_level, token_count, char_count,
			heading, parent_chunk_id, embedding, metadata, created_at, deleted_at
		FROM chunks
		WHERE id = $1 AND deleted_at IS NULL
	`

	chunk := &chunkdomain.Chunk{}
	var metadataJSON []byte
	var embeddingBytes []byte
	var heading sql.NullString
	var parentChunkID *shared.ID

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&chunk.ID, &chunk.ItemVersionID, &chunk.WorkspaceID, &chunk.Content,
		&chunk.ChunkIndex, &chunk.ChunkLevel, &chunk.TokenCount, &chunk.CharCount,
		&heading, &parentChunkID, &embeddingBytes, &metadataJSON,
		&chunk.CreatedAt, &chunk.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrChunkNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get chunk: %w", err)
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

	return chunk, nil
}
