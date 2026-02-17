package getchunk

import (
	"database/sql"

	chunkdomain "documind.jordi.org/internal/retrieval/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

// Handler orchestrates the get-chunk use case.
type Handler struct {
	repo       ReadRepository
	itemReader ItemVersionReader
}

// NewHandler creates a new Handler, wiring the postgres repo internally.
// itemReader can be nil if item enrichment is not needed.
func NewHandler(db *sql.DB, itemReader ItemVersionReader) *Handler {
	return &Handler{
		repo:       newPostgresReadRepo(db),
		itemReader: itemReader,
	}
}

// Handle executes the query and returns the result.
func (h *Handler) Handle(q Query) (*Result, error) {
	chunk, err := h.repo.GetByID(shared.ID(q.ID))
	if err != nil {
		return nil, err
	}

	result := toResult(chunk)

	if q.IncludeItem && h.itemReader != nil {
		title, summary, projectID, status, bodyMd, itemID, err := h.itemReader.GetItemVersionTitle(chunk.ItemVersionID)
		if err == nil {
			result.Item = &ItemResult{
				ID:        string(itemID),
				Title:     title,
				Summary:   summary,
				ProjectID: projectID,
				Status:    status,
				BodyMd:    bodyMd,
			}
		}
	}

	return result, nil
}

// toResult maps a domain chunk to the response Result.
func toResult(chunk *chunkdomain.Chunk) *Result {
	return &Result{
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
