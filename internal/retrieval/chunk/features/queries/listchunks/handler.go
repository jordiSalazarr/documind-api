package listchunks

import (
	"database/sql"

	chunkdomain "documind.jordi.org/internal/retrieval/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

// Handler orchestrates the list-chunks use case.
type Handler struct {
	repo ReadRepository
}

// NewHandler creates a new Handler, wiring the postgres repo internally.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		repo: newPostgresReadRepo(db),
	}
}

// Handle executes the query and returns the results.
func (h *Handler) Handle(q Query) ([]Result, error) {
	q.Limit = clampLimit(q.Limit)
	q.Offset = clampOffset(q.Offset)

	chunks, err := h.repo.ListByItemVersionID(shared.ID(q.ItemVersionID), q.Limit, q.Offset)
	if err != nil {
		return nil, err
	}

	results := make([]Result, len(chunks))
	for i, chunk := range chunks {
		results[i] = toResult(chunk)
	}

	return results, nil
}

// toResult maps a domain chunk to the response Result.
func toResult(chunk *chunkdomain.Chunk) Result {
	return Result{
		ID:            string(chunk.ID),
		Content:       chunk.Content,
		Heading:       chunk.Heading,
		ChunkIndex:    chunk.ChunkIndex,
		TokenCount:    chunk.TokenCount,
		CharCount:     chunk.CharCount,
		ItemVersionID: string(chunk.ItemVersionID),
		Metadata: MetadataResult{
			StartCharOffset: chunk.Metadata.StartCharOffset,
			EndCharOffset:   chunk.Metadata.EndCharOffset,
			CodeBlock:       chunk.Metadata.CodeBlock,
			Language:        chunk.Metadata.Language,
		},
	}
}

func clampLimit(v int) int {
	if v <= 0 {
		return defaultLimit
	}
	if v > maxLimit {
		return maxLimit
	}
	return v
}

func clampOffset(v int) int {
	if v < 0 {
		return 0
	}
	return v
}
