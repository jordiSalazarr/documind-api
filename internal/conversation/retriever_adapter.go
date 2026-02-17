package conversation

import (
	"context"

	"documind.jordi.org/internal/conversation/features/commands/sendmessage"
	"documind.jordi.org/internal/retrieval/search/features/queries/retrieve"
)

// RetrieverAdapter bridges search retrieve.Handler to sendmessage.Retriever interface
type RetrieverAdapter struct {
	handler *retrieve.Handler
}

// NewRetrieverAdapter creates a new retriever adapter
func NewRetrieverAdapter(handler *retrieve.Handler) *RetrieverAdapter {
	return &RetrieverAdapter{handler: handler}
}

// Retrieve translates between the chat and search types
func (a *RetrieverAdapter) Retrieve(ctx context.Context, input sendmessage.RetrievalInput) (*sendmessage.RetrievalOutput, error) {
	result, err := a.handler.Handle(ctx, retrieve.Input{
		Query:       input.Query,
		WorkspaceID: input.WorkspaceID,
		MaxResults:  input.MaxResults,
		MaxTokens:   input.MaxTokens,
	})
	if err != nil {
		return nil, err
	}

	output := &sendmessage.RetrievalOutput{
		Results: make([]*sendmessage.RetrievalResult, len(result.Results)),
	}
	for i, r := range result.Results {
		output.Results[i] = &sendmessage.RetrievalResult{
			ChunkID:       r.Chunk.ID,
			ItemVersionID: r.Chunk.ItemVersionID,
			Heading:       r.Chunk.Heading,
			Content:       r.Chunk.Content,
			TokenCount:    r.Chunk.TokenCount,
			FinalScore:    r.FinalScore,
		}
	}

	return output, nil
}
