package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/pgvector/pgvector-go"

	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
)

type ChunkRepository struct {
	db *sql.DB
}

func NewChunkRepository(db *sql.DB) *ChunkRepository {
	return &ChunkRepository{db: db}
}

// Create creates a new chunk
func (r *ChunkRepository) Create(chunk *knowledge.Chunk) error {
	return r.CreateWithContext(context.Background(), chunk)
}

// CreateWithContext creates a new chunk with context
func (r *ChunkRepository) CreateWithContext(ctx context.Context, chunk *knowledge.Chunk) error {
	metadataJSON, err := json.Marshal(chunk.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	var embedding interface{}
	if chunk.Embedding != nil {
		embedding = pgvector.NewVector(chunk.Embedding)
	}

	query := `
		INSERT INTO chunks (
			id, item_version_id, workspace_id, content,
			chunk_index, chunk_level, token_count, char_count,
			heading, parent_chunk_id, embedding, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = r.db.ExecContext(ctx, query,
		chunk.ID, chunk.ItemVersionID, chunk.WorkspaceID, chunk.Content,
		chunk.ChunkIndex, chunk.ChunkLevel, chunk.TokenCount, chunk.CharCount,
		nullString(chunk.Heading), chunk.ParentChunkID, embedding,
		metadataJSON, chunk.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create chunk: %w", err)
	}

	return nil
}

// CreateBatch creates multiple chunks in a single transaction
func (r *ChunkRepository) CreateBatch(chunks []*knowledge.Chunk) error {
	return r.CreateBatchWithContext(context.Background(), chunks)
}

// CreateBatchWithContext creates multiple chunks in a single transaction with context
func (r *ChunkRepository) CreateBatchWithContext(ctx context.Context, chunks []*knowledge.Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

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

// GetByID retrieves a chunk by ID
func (r *ChunkRepository) GetByID(id common.ID) (*knowledge.Chunk, error) {
	return r.GetByIDWithContext(context.Background(), id)
}

// GetByIDWithContext retrieves a chunk by ID with context
func (r *ChunkRepository) GetByIDWithContext(ctx context.Context, id common.ID) (*knowledge.Chunk, error) {
	query := `
		SELECT id, item_version_id, workspace_id, content,
			chunk_index, chunk_level, token_count, char_count,
			heading, parent_chunk_id, embedding, metadata, created_at, deleted_at
		FROM chunks
		WHERE id = $1 AND deleted_at IS NULL
	`

	chunk := &knowledge.Chunk{}
	var metadataJSON []byte
	var embeddingBytes []byte
	var heading sql.NullString
	var parentChunkID *common.ID

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&chunk.ID, &chunk.ItemVersionID, &chunk.WorkspaceID, &chunk.Content,
		&chunk.ChunkIndex, &chunk.ChunkLevel, &chunk.TokenCount, &chunk.CharCount,
		&heading, &parentChunkID, &embeddingBytes, &metadataJSON,
		&chunk.CreatedAt, &chunk.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("chunk not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get chunk: %w", err)
	}

	// Parse optional fields
	if heading.Valid {
		chunk.Heading = heading.String
	}
	chunk.ParentChunkID = parentChunkID

	// Parse embedding if present
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

// ListByItemVersionID retrieves all chunks for an item version
func (r *ChunkRepository) ListByItemVersionID(itemVersionID common.ID) ([]*knowledge.Chunk, error) {
	return r.ListByItemVersionIDWithContext(context.Background(), itemVersionID)
}

// ListByItemVersionIDWithContext retrieves all chunks for an item version with context
func (r *ChunkRepository) ListByItemVersionIDWithContext(ctx context.Context, itemVersionID common.ID) ([]*knowledge.Chunk, error) {
	query := `
		SELECT id, item_version_id, workspace_id, content,
			chunk_index, chunk_level, token_count, char_count,
			heading, parent_chunk_id, embedding, metadata, created_at, deleted_at
		FROM chunks
		WHERE item_version_id = $1 AND deleted_at IS NULL
		ORDER BY chunk_index ASC
	`

	rows, err := r.db.QueryContext(ctx, query, itemVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list chunks: %w", err)
	}
	defer rows.Close()

	chunks := []*knowledge.Chunk{}
	for rows.Next() {
		chunk := &knowledge.Chunk{}
		var metadataJSON []byte
		var embeddingBytes []byte
		var heading sql.NullString
		var parentChunkID *common.ID

		err := rows.Scan(
			&chunk.ID, &chunk.ItemVersionID, &chunk.WorkspaceID, &chunk.Content,
			&chunk.ChunkIndex, &chunk.ChunkLevel, &chunk.TokenCount, &chunk.CharCount,
			&heading, &parentChunkID, &embeddingBytes, &metadataJSON,
			&chunk.CreatedAt, &chunk.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chunk: %w", err)
		}

		// Parse optional fields
		if heading.Valid {
			chunk.Heading = heading.String
		}
		chunk.ParentChunkID = parentChunkID

		// Parse embedding if present
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

// UpdateEmbedding updates the embedding for a chunk
func (r *ChunkRepository) UpdateEmbedding(id common.ID, embedding []float32) error {
	return r.UpdateEmbeddingWithContext(context.Background(), id, embedding)
}

// UpdateEmbeddingWithContext updates the embedding for a chunk with context
func (r *ChunkRepository) UpdateEmbeddingWithContext(ctx context.Context, id common.ID, embedding []float32) error {
	query := `UPDATE chunks SET embedding = $1 WHERE id = $2`

	vec := pgvector.NewVector(embedding)
	_, err := r.db.ExecContext(ctx, query, vec, id)
	if err != nil {
		return fmt.Errorf("failed to update embedding: %w", err)
	}

	return nil
}

// UpdateEmbeddingsBatch updates embeddings for multiple chunks
func (r *ChunkRepository) UpdateEmbeddingsBatch(updates map[common.ID][]float32) error {
	return r.UpdateEmbeddingsBatchWithContext(context.Background(), updates)
}

// UpdateEmbeddingsBatchWithContext updates embeddings for multiple chunks with context
func (r *ChunkRepository) UpdateEmbeddingsBatchWithContext(ctx context.Context, updates map[common.ID][]float32) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
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

// SearchByEmbedding performs vector similarity search
func (r *ChunkRepository) SearchByEmbedding(
	workspaceID common.ID,
	embedding []float32,
	limit int,
	minSimilarity float64,
) ([]*knowledge.Chunk, error) {
	return r.SearchByEmbeddingWithContext(context.Background(), workspaceID, embedding, limit, minSimilarity)
}

// SearchByEmbeddingWithContext performs vector similarity search with context
func (r *ChunkRepository) SearchByEmbeddingWithContext(
	ctx context.Context,
	workspaceID common.ID,
	embedding []float32,
	limit int,
	minSimilarity float64,
) ([]*knowledge.Chunk, error) {
	// Use cosine distance: distance = 1 - similarity
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
	rows, err := r.db.QueryContext(ctx, query, vec, workspaceID, maxDistance, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search by embedding: %w", err)
	}
	defer rows.Close()

	chunks := []*knowledge.Chunk{}
	for rows.Next() {
		chunk := &knowledge.Chunk{}
		var metadataJSON []byte
		var embeddingVec pgvector.Vector
		var heading sql.NullString
		var parentChunkID *common.ID
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

		// Parse optional fields
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

// SearchByText performs full-text search on chunk content
func (r *ChunkRepository) SearchByText(workspaceID common.ID, query string, limit int) ([]*knowledge.Chunk, error) {
	return r.SearchByTextWithContext(context.Background(), workspaceID, query, limit)
}

// SearchByTextWithContext performs full-text search on chunk content with context
func (r *ChunkRepository) SearchByTextWithContext(ctx context.Context, workspaceID common.ID, query string, limit int) ([]*knowledge.Chunk, error) {
	sql := `
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

	rows, err := r.db.QueryContext(ctx, sql, query, workspaceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search by text: %w", err)
	}
	defer rows.Close()

	chunks := []*knowledge.Chunk{}
	for rows.Next() {
		chunk := &knowledge.Chunk{}
		var metadataJSON []byte
		var embeddingBytes []byte
		var heading null.String
		var parentChunkID *common.ID
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

		// Parse optional fields
		if heading.Valid {
			chunk.Heading = heading.String
		}
		chunk.ParentChunkID = parentChunkID

		// Parse embedding if present
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

// DeleteByItemVersionID deletes all chunks for an item version (soft delete)
func (r *ChunkRepository) DeleteByItemVersionID(itemVersionID common.ID) error {
	return r.DeleteByItemVersionIDWithContext(context.Background(), itemVersionID)
}

// DeleteByItemVersionIDWithContext deletes all chunks for an item version with context
func (r *ChunkRepository) DeleteByItemVersionIDWithContext(ctx context.Context, itemVersionID common.ID) error {
	query := `UPDATE chunks SET deleted_at = NOW() WHERE item_version_id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, itemVersionID)
	if err != nil {
		return fmt.Errorf("failed to delete chunks: %w", err)
	}

	return nil
}

// CountByItemVersionID counts chunks for an item version
func (r *ChunkRepository) CountByItemVersionID(itemVersionID common.ID) (int, error) {
	return r.CountByItemVersionIDWithContext(context.Background(), itemVersionID)
}

// CountByItemVersionIDWithContext counts chunks for an item version with context
func (r *ChunkRepository) CountByItemVersionIDWithContext(ctx context.Context, itemVersionID common.ID) (int, error) {
	query := `SELECT COUNT(*) FROM chunks WHERE item_version_id = $1 AND deleted_at IS NULL`

	var count int
	err := r.db.QueryRowContext(ctx, query, itemVersionID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count chunks: %w", err)
	}

	return count, nil
}

// Helper function to convert string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
