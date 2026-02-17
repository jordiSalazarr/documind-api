package retrieval

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	retrievaldomain "documind.jordi.org/internal/retrieval/domain"
	"documind.jordi.org/internal/shared/infrastructure/openai"
)

// Reranker scores and re-orders retrieval results by semantic relevance to the query.
type Reranker interface {
	Rerank(ctx context.Context, query string, results []*retrievaldomain.RetrievalResult, topK int) ([]*retrievaldomain.RetrievalResult, error)
}

// LLMReranker uses an LLM (GPT-4o-mini) to score (query, document) pairs for relevance.
type LLMReranker struct {
	chatClient *openai.ChatClient
}

// NewLLMReranker creates a new LLM-based reranker.
func NewLLMReranker(chatClient *openai.ChatClient) *LLMReranker {
	return &LLMReranker{chatClient: chatClient}
}

// Rerank scores each result against the query using the LLM and re-sorts by relevance.
// Falls back to the original RRF ordering on any error.
func (r *LLMReranker) Rerank(
	ctx context.Context,
	query string,
	results []*retrievaldomain.RetrievalResult,
	topK int,
) ([]*retrievaldomain.RetrievalResult, error) {
	if len(results) == 0 {
		return results, nil
	}

	// Build the batch scoring prompt
	var docs strings.Builder
	for i, result := range results {
		content := result.Chunk.Content
		// Truncate long chunks to keep prompt manageable
		if len(content) > 500 {
			content = content[:500] + "..."
		}
		docs.WriteString(fmt.Sprintf("[%d] %s\n\n", i+1, content))
	}

	prompt := fmt.Sprintf(`Score the relevance of each document to the query on a scale of 0-10.
Return ONLY a JSON array of integer scores in the same order as the documents.
Example response: [8, 3, 9, 1, 5]

Query: "%s"

Documents:
%s
Scores:`, query, docs.String())

	messages := []openai.ChatMessage{
		{
			Role:    "system",
			Content: "You are a relevance scoring system. You score documents by their relevance to a query. Respond only with a JSON array of integer scores.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	resp, err := r.chatClient.CreateCompletion(ctx, messages, 100)
	if err != nil {
		// Fall back to original ordering
		return fallbackTopK(results, topK), nil
	}

	if len(resp.Choices) == 0 {
		return fallbackTopK(results, topK), nil
	}

	// Parse scores from LLM response
	scores, err := parseScores(resp.Choices[0].Message.Content, len(results))
	if err != nil {
		return fallbackTopK(results, topK), nil
	}

	// Apply rerank scores and re-sort
	for i, result := range results {
		if i < len(scores) {
			result.RerankScore = scores[i]
			result.FinalScore = scores[i]
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].RerankScore > results[j].RerankScore
	})

	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

// parseScores extracts a JSON array of float64 scores from the LLM response.
func parseScores(response string, expectedCount int) ([]float64, error) {
	response = strings.TrimSpace(response)

	// Try direct JSON parse first
	var scores []float64
	if err := json.Unmarshal([]byte(response), &scores); err == nil {
		if len(scores) >= expectedCount {
			return scores[:expectedCount], nil
		}
		return scores, nil
	}

	// Try extracting JSON array from response text
	start := strings.Index(response, "[")
	end := strings.LastIndex(response, "]")
	if start >= 0 && end > start {
		jsonStr := response[start : end+1]
		if err := json.Unmarshal([]byte(jsonStr), &scores); err == nil {
			if len(scores) >= expectedCount {
				return scores[:expectedCount], nil
			}
			return scores, nil
		}
	}

	// Last resort: try to parse comma-separated numbers
	parts := strings.Split(strings.Trim(response, "[] \n"), ",")
	scores = make([]float64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if val, err := strconv.ParseFloat(p, 64); err == nil {
			scores = append(scores, val)
		}
	}

	if len(scores) == 0 {
		return nil, fmt.Errorf("could not parse scores from response: %s", response)
	}

	return scores, nil
}

// fallbackTopK returns the top-K results based on original RRF ordering.
func fallbackTopK(results []*retrievaldomain.RetrievalResult, topK int) []*retrievaldomain.RetrievalResult {
	if len(results) > topK {
		return results[:topK]
	}
	return results
}
