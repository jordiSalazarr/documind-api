package retrieve

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	retrievaldomain "documind.jordi.org/internal/retrieval/domain"
	"documind.jordi.org/internal/retrieval"
	shared "documind.jordi.org/internal/shared/domain"
	"documind.jordi.org/internal/shared/infrastructure/openai"
)

// Handler handles hybrid retrieval operations
type Handler struct {
	chunkRepo       ChunkSearchRepository
	embeddingClient *openai.EmbeddingClient
	queryProcessor  *retrieval.QueryProcessor
	reranker        Reranker // optional LLM-based reranker
	config          retrievaldomain.RetrievalConfig
}

// NewHandler creates a new retrieval handler
func NewHandler(
	chunkRepo ChunkSearchRepository,
	embeddingClient *openai.EmbeddingClient,
	config retrievaldomain.RetrievalConfig,
) *Handler {
	// Apply defaults if not set
	if config.VectorTopK <= 0 {
		config.VectorTopK = 50
	}
	if config.FullTextTopK <= 0 {
		config.FullTextTopK = 50
	}
	if config.RRFConstant <= 0 {
		config.RRFConstant = 60
	}
	if config.FinalTopK <= 0 {
		config.FinalTopK = 5
	}
	if config.MinSimilarity <= 0 {
		config.MinSimilarity = 0.5
	}
	if config.VectorWeight <= 0 {
		config.VectorWeight = 0.7
	}
	if config.FullTextWeight <= 0 {
		config.FullTextWeight = 0.3
	}
	if config.RerankTopK <= 0 {
		config.RerankTopK = 20
	}

	// Initialize query processor with defaults
	queryProcessor := retrieval.NewQueryProcessor(retrieval.DefaultQueryProcessorConfig())

	return &Handler{
		chunkRepo:       chunkRepo,
		embeddingClient: embeddingClient,
		queryProcessor:  queryProcessor,
		config:          config,
	}
}

// SetReranker sets an optional reranker for cross-encoder re-ranking.
func (h *Handler) SetReranker(reranker Reranker) {
	h.reranker = reranker
}

// Handle performs hybrid retrieval combining vector and full-text search
func (h *Handler) Handle(ctx context.Context, input Input) (*Output, error) {
	startTime := time.Now()
	stats := &retrievaldomain.RetrievalStats{}

	// Apply defaults
	maxResults := input.MaxResults
	if maxResults <= 0 {
		maxResults = h.config.FinalTopK
	}

	// Process query for normalization, expansion, and intent detection
	processedQuery := h.queryProcessor.Process(input.Query)

	// Determine which query to use for full-text search
	// Use expanded query for better recall if expansion is enabled
	fullTextQuery := processedQuery.Normalized
	if input.EnableQueryExpand || h.queryProcessor.EnableExpansion() {
		fullTextQuery = processedQuery.Expanded
	}

	// Generate query embedding using the normalized query
	queryEmbedding, err := h.embeddingClient.CreateEmbedding(ctx, processedQuery.Normalized)
	if err != nil {
		// Fall back to text-only search if embedding fails
		queryEmbedding = nil
	}

	// Run vector and full-text search in parallel
	var (
		vectorResults   []*retrievaldomain.Chunk
		fulltextResults []*retrievaldomain.Chunk
		vectorErr       error
		fulltextErr     error
		wg              sync.WaitGroup
	)

	// Vector search
	if queryEmbedding != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			vectorResults, vectorErr = h.chunkRepo.SearchByEmbedding(
				input.WorkspaceID,
				queryEmbedding,
				h.config.VectorTopK,
				h.config.MinSimilarity,
			)
			stats.VectorSearchTimeMs = time.Since(start).Milliseconds()
		}()
	}

	// Full-text search with processed/expanded query
	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		fulltextResults, fulltextErr = h.chunkRepo.SearchByText(
			input.WorkspaceID,
			fullTextQuery,
			h.config.FullTextTopK,
		)
		stats.FullTextSearchTimeMs = time.Since(start).Milliseconds()
	}()

	wg.Wait()

	// Handle errors (log but continue with available results)
	if vectorErr != nil {
		vectorResults = []*retrievaldomain.Chunk{}
	}
	if fulltextErr != nil {
		fulltextResults = []*retrievaldomain.Chunk{}
	}

	stats.TotalVectorResults = len(vectorResults)
	stats.TotalFullTextResults = len(fulltextResults)

	// Combine results using Reciprocal Rank Fusion
	fusedResults := h.reciprocalRankFusion(vectorResults, fulltextResults, queryEmbedding)
	stats.TotalUniqueResults = len(fusedResults)

	// Apply cross-encoder re-ranking if enabled
	if h.config.EnableReranking && h.reranker != nil && len(fusedResults) > 0 {
		rerankCandidates := fusedResults
		if len(rerankCandidates) > h.config.RerankTopK {
			rerankCandidates = rerankCandidates[:h.config.RerankTopK]
		}
		rerankStart := time.Now()
		reranked, err := h.reranker.Rerank(ctx, processedQuery.Normalized, rerankCandidates, maxResults)
		stats.RerankTimeMs = time.Since(rerankStart).Milliseconds()
		if err == nil {
			fusedResults = reranked
		}
	}

	// Apply MMR for diversity if enabled
	if h.config.EnableMMR && len(fusedResults) > maxResults {
		fusedResults = h.applyMMR(fusedResults, queryEmbedding, maxResults)
	}

	// Limit to requested number of results
	if len(fusedResults) > maxResults {
		fusedResults = fusedResults[:maxResults]
	}

	// Apply token limit if specified
	if input.MaxTokens > 0 {
		fusedResults = h.applyTokenLimit(fusedResults, input.MaxTokens)
	}

	stats.TotalTimeMs = time.Since(startTime).Milliseconds()

	return &Output{
		Results:        fusedResults,
		Stats:          stats,
		ProcessedQuery: processedQuery,
	}, nil
}

// HandleWithMultiQuery performs retrieval using multiple query variations
// This can improve recall for complex or ambiguous queries
func (h *Handler) HandleWithMultiQuery(ctx context.Context, input Input) (*Output, error) {
	// Generate multiple query variations
	queryVariations := h.queryProcessor.GenerateMultipleQueries(input.Query)

	if len(queryVariations) <= 1 {
		// Only one query, use standard retrieval
		return h.Handle(ctx, input)
	}

	startTime := time.Now()
	stats := &retrievaldomain.RetrievalStats{}

	// Process the original query for metadata
	processedQuery := h.queryProcessor.Process(input.Query)

	// Collect results from all query variations
	allResultsMap := make(map[shared.ID]*retrievaldomain.RetrievalResult)
	queryScores := make(map[shared.ID][]float64) // Track scores across queries

	for _, queryVar := range queryVariations {
		varInput := Input{
			Query:             queryVar,
			WorkspaceID:       input.WorkspaceID,
			ProjectID:         input.ProjectID,
			MaxResults:        h.config.VectorTopK, // Get more results per query
			EnableQueryExpand: false,                // Already expanded
		}

		result, err := h.Handle(ctx, varInput)
		if err != nil {
			continue // Skip failed queries
		}

		// Merge results
		for _, r := range result.Results {
			chunkID := r.Chunk.ID
			if existing, ok := allResultsMap[chunkID]; ok {
				// Combine scores - boost items that appear in multiple queries
				existing.RRFScore += r.RRFScore
				queryScores[chunkID] = append(queryScores[chunkID], r.RRFScore)
			} else {
				allResultsMap[chunkID] = r
				queryScores[chunkID] = []float64{r.RRFScore}
			}
		}

		// Accumulate timing stats
		if result.Stats != nil {
			stats.VectorSearchTimeMs += result.Stats.VectorSearchTimeMs
			stats.FullTextSearchTimeMs += result.Stats.FullTextSearchTimeMs
			stats.TotalVectorResults += result.Stats.TotalVectorResults
			stats.TotalFullTextResults += result.Stats.TotalFullTextResults
		}
	}

	// Apply query frequency bonus - boost results that appear in multiple queries
	for chunkID, result := range allResultsMap {
		scores := queryScores[chunkID]
		queryFrequency := float64(len(scores)) / float64(len(queryVariations))
		// Boost score by query frequency (results in more queries get higher scores)
		result.FinalScore = result.RRFScore * (1 + queryFrequency*0.5)
	}

	// Convert to slice and sort
	results := make([]*retrievaldomain.RetrievalResult, 0, len(allResultsMap))
	for _, result := range allResultsMap {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].FinalScore > results[j].FinalScore
	})

	stats.TotalUniqueResults = len(results)

	// Apply final limits
	maxResults := input.MaxResults
	if maxResults <= 0 {
		maxResults = h.config.FinalTopK
	}
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	if input.MaxTokens > 0 {
		results = h.applyTokenLimit(results, input.MaxTokens)
	}

	stats.TotalTimeMs = time.Since(startTime).Milliseconds()

	return &Output{
		Results:        results,
		Stats:          stats,
		ProcessedQuery: processedQuery,
	}, nil
}

// RetrieveByVector performs vector-only search
func (h *Handler) RetrieveByVector(
	ctx context.Context,
	workspaceID shared.ID,
	queryEmbedding []float32,
	topK int,
	minSimilarity float64,
) ([]*retrievaldomain.RetrievalResult, error) {
	if topK <= 0 {
		topK = h.config.VectorTopK
	}
	if minSimilarity <= 0 {
		minSimilarity = h.config.MinSimilarity
	}

	chunks, err := h.chunkRepo.SearchByEmbedding(workspaceID, queryEmbedding, topK, minSimilarity)
	if err != nil {
		return nil, err
	}

	results := make([]*retrievaldomain.RetrievalResult, len(chunks))
	for i, chunk := range chunks {
		similarity := computeVectorSimilarity(chunk.Embedding, queryEmbedding)
		results[i] = &retrievaldomain.RetrievalResult{
			Chunk:           chunk,
			VectorRank:      i + 1,
			VectorScore:     similarity,
			FinalScore:      similarity,
			RetrievalMethod: "vector",
		}
	}

	return results, nil
}

// RetrieveByText performs full-text-only search
func (h *Handler) RetrieveByText(
	ctx context.Context,
	workspaceID shared.ID,
	query string,
	topK int,
) ([]*retrievaldomain.RetrievalResult, error) {
	if topK <= 0 {
		topK = h.config.FullTextTopK
	}

	chunks, err := h.chunkRepo.SearchByText(workspaceID, query, topK)
	if err != nil {
		return nil, err
	}

	results := make([]*retrievaldomain.RetrievalResult, len(chunks))
	for i, chunk := range chunks {
		score := 1.0 / float64(i+1) // Simple rank-based score
		results[i] = &retrievaldomain.RetrievalResult{
			Chunk:           chunk,
			FullTextRank:    i + 1,
			FullTextScore:   score,
			FinalScore:      score,
			RetrievalMethod: "fulltext",
		}
	}

	return results, nil
}

// GetConfig returns the current retrieval configuration
func (h *Handler) GetConfig() retrievaldomain.RetrievalConfig {
	return h.config
}

// GetQueryProcessor returns the query processor for external use
func (h *Handler) GetQueryProcessor() *retrieval.QueryProcessor {
	return h.queryProcessor
}

// reciprocalRankFusion combines results from multiple retrieval methods using RRF
func (h *Handler) reciprocalRankFusion(
	vectorResults []*retrievaldomain.Chunk,
	fulltextResults []*retrievaldomain.Chunk,
	queryEmbedding []float32,
) []*retrievaldomain.RetrievalResult {
	k := h.config.RRFConstant
	resultMap := make(map[shared.ID]*retrievaldomain.RetrievalResult)

	// Process vector results
	for rank, chunk := range vectorResults {
		id := chunk.ID
		rrfScore := 1.0 / float64(rank+k)

		if result, exists := resultMap[id]; exists {
			result.VectorRank = rank + 1
			result.VectorScore = computeVectorSimilarity(chunk.Embedding, queryEmbedding)
			result.RRFScore += rrfScore
			result.RetrievalMethod = "hybrid"
		} else {
			resultMap[id] = &retrievaldomain.RetrievalResult{
				Chunk:           chunk,
				VectorRank:      rank + 1,
				VectorScore:     computeVectorSimilarity(chunk.Embedding, queryEmbedding),
				RRFScore:        rrfScore,
				RetrievalMethod: "vector",
			}
		}
	}

	// Process full-text results
	for rank, chunk := range fulltextResults {
		id := chunk.ID
		rrfScore := 1.0 / float64(rank+k)

		if result, exists := resultMap[id]; exists {
			result.FullTextRank = rank + 1
			result.FullTextScore = 1.0 / float64(rank+1) // Normalize rank to score
			result.RRFScore += rrfScore
			result.RetrievalMethod = "hybrid"
		} else {
			resultMap[id] = &retrievaldomain.RetrievalResult{
				Chunk:           chunk,
				FullTextRank:    rank + 1,
				FullTextScore:   1.0 / float64(rank+1),
				RRFScore:        rrfScore,
				RetrievalMethod: "fulltext",
			}
		}
	}

	// Convert map to slice
	results := make([]*retrievaldomain.RetrievalResult, 0, len(resultMap))
	for _, result := range resultMap {
		// Compute final score as weighted combination
		result.FinalScore = result.RRFScore
		results = append(results, result)
	}

	// Sort by RRF score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].RRFScore > results[j].RRFScore
	})

	return results
}

// computeVectorSimilarity calculates cosine similarity between two vectors
func computeVectorSimilarity(embedding1, embedding2 []float32) float64 {
	if len(embedding1) == 0 || len(embedding2) == 0 {
		return 0
	}
	if len(embedding1) != len(embedding2) {
		return 0
	}

	var dotProduct, norm1, norm2 float64
	for i := 0; i < len(embedding1); i++ {
		dotProduct += float64(embedding1[i]) * float64(embedding2[i])
		norm1 += float64(embedding1[i]) * float64(embedding1[i])
		norm2 += float64(embedding2[i]) * float64(embedding2[i])
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// applyMMR applies Maximal Marginal Relevance for diversity
func (h *Handler) applyMMR(
	results []*retrievaldomain.RetrievalResult,
	_ []float32, // queryEmbedding - reserved for future use with query-based relevance
	topK int,
) []*retrievaldomain.RetrievalResult {
	if len(results) == 0 {
		return results
	}

	lambda := h.config.DiversityLambda
	selected := make([]*retrievaldomain.RetrievalResult, 0, topK)
	remaining := make([]*retrievaldomain.RetrievalResult, len(results))
	copy(remaining, results)

	for len(selected) < topK && len(remaining) > 0 {
		maxScore := math.Inf(-1)
		maxIdx := 0

		for i, candidate := range remaining {
			// Relevance score (using RRF score as proxy)
			relevance := candidate.RRFScore

			// Compute max similarity to already selected items
			maxSim := 0.0
			for _, sel := range selected {
				if candidate.Chunk.Embedding != nil && sel.Chunk.Embedding != nil {
					sim := computeVectorSimilarity(candidate.Chunk.Embedding, sel.Chunk.Embedding)
					if sim > maxSim {
						maxSim = sim
					}
				}
			}

			// MMR score: lambda * relevance - (1-lambda) * max_similarity
			mmrScore := lambda*relevance - (1-lambda)*maxSim

			if mmrScore > maxScore {
				maxScore = mmrScore
				maxIdx = i
			}
		}

		// Add best candidate to selected
		selected = append(selected, remaining[maxIdx])
		// Remove from remaining
		remaining = append(remaining[:maxIdx], remaining[maxIdx+1:]...)
	}

	return selected
}

// applyTokenLimit limits results to fit within a token budget
func (h *Handler) applyTokenLimit(
	results []*retrievaldomain.RetrievalResult,
	maxTokens int,
) []*retrievaldomain.RetrievalResult {
	totalTokens := 0
	limited := make([]*retrievaldomain.RetrievalResult, 0, len(results))

	for _, result := range results {
		if totalTokens+result.Chunk.TokenCount > maxTokens {
			break
		}
		limited = append(limited, result)
		totalTokens += result.Chunk.TokenCount
	}

	return limited
}
