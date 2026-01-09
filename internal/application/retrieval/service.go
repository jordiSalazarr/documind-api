package retrieval

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
	"documind.jordi.org/internal/infrastructure/openai"
)

// Service handles hybrid retrieval operations
type Service struct {
	chunkRepo       knowledge.ChunkRepository
	embeddingClient *openai.EmbeddingClient
	queryProcessor  *QueryProcessor
	config          knowledge.RetrievalConfig
}

// NewService creates a new retrieval service
func NewService(
	chunkRepo knowledge.ChunkRepository,
	embeddingClient *openai.EmbeddingClient,
	config knowledge.RetrievalConfig,
) *Service {
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

	// Initialize query processor with defaults
	queryProcessor := NewQueryProcessor(DefaultQueryProcessorConfig())

	return &Service{
		chunkRepo:       chunkRepo,
		embeddingClient: embeddingClient,
		queryProcessor:  queryProcessor,
		config:          config,
	}
}

// RetrieveInput contains parameters for hybrid retrieval
type RetrieveInput struct {
	Query              string
	WorkspaceID        common.ID
	ProjectID          *common.ID
	MaxResults         int
	MaxTokens          int  // Max total tokens across all results
	EnableQueryExpand  bool // Enable query expansion (default: true)
	EnableMultiQuery   bool // Enable multi-query retrieval for better recall
}

// RetrieveResult contains the retrieval results
type RetrieveResult struct {
	Results        []*knowledge.RetrievalResult
	Stats          *knowledge.RetrievalStats
	ProcessedQuery *ProcessedQuery // Query processing information
}

// Retrieve performs hybrid retrieval combining vector and full-text search
func (s *Service) Retrieve(ctx context.Context, input RetrieveInput) (*RetrieveResult, error) {
	startTime := time.Now()
	stats := &knowledge.RetrievalStats{}

	// Apply defaults
	maxResults := input.MaxResults
	if maxResults <= 0 {
		maxResults = s.config.FinalTopK
	}

	// Process query for normalization, expansion, and intent detection
	processedQuery := s.queryProcessor.Process(input.Query)

	// Determine which query to use for full-text search
	// Use expanded query for better recall if expansion is enabled
	fullTextQuery := processedQuery.Normalized
	if input.EnableQueryExpand || s.queryProcessor.enableExpansion {
		fullTextQuery = processedQuery.Expanded
	}

	// Generate query embedding using the normalized query
	queryEmbedding, err := s.embeddingClient.CreateEmbedding(ctx, processedQuery.Normalized)
	if err != nil {
		// Fall back to text-only search if embedding fails
		queryEmbedding = nil
	}

	// Run vector and full-text search in parallel
	var (
		vectorResults   []*knowledge.Chunk
		fulltextResults []*knowledge.Chunk
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
			vectorResults, vectorErr = s.chunkRepo.SearchByEmbedding(
				input.WorkspaceID,
				queryEmbedding,
				s.config.VectorTopK,
				s.config.MinSimilarity,
			)
			stats.VectorSearchTimeMs = time.Since(start).Milliseconds()
		}()
	}

	// Full-text search with processed/expanded query
	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		fulltextResults, fulltextErr = s.chunkRepo.SearchByText(
			input.WorkspaceID,
			fullTextQuery,
			s.config.FullTextTopK,
		)
		stats.FullTextSearchTimeMs = time.Since(start).Milliseconds()
	}()

	wg.Wait()

	// Handle errors (log but continue with available results)
	if vectorErr != nil {
		vectorResults = []*knowledge.Chunk{}
	}
	if fulltextErr != nil {
		fulltextResults = []*knowledge.Chunk{}
	}

	stats.TotalVectorResults = len(vectorResults)
	stats.TotalFullTextResults = len(fulltextResults)

	// Combine results using Reciprocal Rank Fusion
	fusedResults := s.reciprocalRankFusion(vectorResults, fulltextResults, queryEmbedding)
	stats.TotalUniqueResults = len(fusedResults)

	// Apply MMR for diversity if enabled
	if s.config.EnableMMR && len(fusedResults) > maxResults {
		fusedResults = s.applyMMR(fusedResults, queryEmbedding, maxResults)
	}

	// Limit to requested number of results
	if len(fusedResults) > maxResults {
		fusedResults = fusedResults[:maxResults]
	}

	// Apply token limit if specified
	if input.MaxTokens > 0 {
		fusedResults = s.applyTokenLimit(fusedResults, input.MaxTokens)
	}

	stats.TotalTimeMs = time.Since(startTime).Milliseconds()

	return &RetrieveResult{
		Results:        fusedResults,
		Stats:          stats,
		ProcessedQuery: processedQuery,
	}, nil
}

// reciprocalRankFusion combines results from multiple retrieval methods using RRF
func (s *Service) reciprocalRankFusion(
	vectorResults []*knowledge.Chunk,
	fulltextResults []*knowledge.Chunk,
	queryEmbedding []float32,
) []*knowledge.RetrievalResult {
	k := s.config.RRFConstant
	resultMap := make(map[common.ID]*knowledge.RetrievalResult)

	// Process vector results
	for rank, chunk := range vectorResults {
		id := chunk.ID
		rrfScore := 1.0 / float64(rank+k)

		if result, exists := resultMap[id]; exists {
			result.VectorRank = rank + 1
			result.VectorScore = s.computeVectorSimilarity(chunk.Embedding, queryEmbedding)
			result.RRFScore += rrfScore
			result.RetrievalMethod = "hybrid"
		} else {
			resultMap[id] = &knowledge.RetrievalResult{
				Chunk:           chunk,
				VectorRank:      rank + 1,
				VectorScore:     s.computeVectorSimilarity(chunk.Embedding, queryEmbedding),
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
			resultMap[id] = &knowledge.RetrievalResult{
				Chunk:           chunk,
				FullTextRank:    rank + 1,
				FullTextScore:   1.0 / float64(rank+1),
				RRFScore:        rrfScore,
				RetrievalMethod: "fulltext",
			}
		}
	}

	// Convert map to slice
	results := make([]*knowledge.RetrievalResult, 0, len(resultMap))
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
func (s *Service) computeVectorSimilarity(embedding1, embedding2 []float32) float64 {
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
func (s *Service) applyMMR(
	results []*knowledge.RetrievalResult,
	_ []float32, // queryEmbedding - reserved for future use with query-based relevance
	topK int,
) []*knowledge.RetrievalResult {
	if len(results) == 0 {
		return results
	}

	lambda := s.config.DiversityLambda
	selected := make([]*knowledge.RetrievalResult, 0, topK)
	remaining := make([]*knowledge.RetrievalResult, len(results))
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
					sim := s.computeVectorSimilarity(candidate.Chunk.Embedding, sel.Chunk.Embedding)
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
func (s *Service) applyTokenLimit(
	results []*knowledge.RetrievalResult,
	maxTokens int,
) []*knowledge.RetrievalResult {
	totalTokens := 0
	limited := make([]*knowledge.RetrievalResult, 0, len(results))

	for _, result := range results {
		if totalTokens+result.Chunk.TokenCount > maxTokens {
			break
		}
		limited = append(limited, result)
		totalTokens += result.Chunk.TokenCount
	}

	return limited
}

// RetrieveByVector performs vector-only search
func (s *Service) RetrieveByVector(
	ctx context.Context,
	workspaceID common.ID,
	queryEmbedding []float32,
	topK int,
	minSimilarity float64,
) ([]*knowledge.RetrievalResult, error) {
	if topK <= 0 {
		topK = s.config.VectorTopK
	}
	if minSimilarity <= 0 {
		minSimilarity = s.config.MinSimilarity
	}

	chunks, err := s.chunkRepo.SearchByEmbedding(workspaceID, queryEmbedding, topK, minSimilarity)
	if err != nil {
		return nil, err
	}

	results := make([]*knowledge.RetrievalResult, len(chunks))
	for i, chunk := range chunks {
		similarity := s.computeVectorSimilarity(chunk.Embedding, queryEmbedding)
		results[i] = &knowledge.RetrievalResult{
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
func (s *Service) RetrieveByText(
	ctx context.Context,
	workspaceID common.ID,
	query string,
	topK int,
) ([]*knowledge.RetrievalResult, error) {
	if topK <= 0 {
		topK = s.config.FullTextTopK
	}

	chunks, err := s.chunkRepo.SearchByText(workspaceID, query, topK)
	if err != nil {
		return nil, err
	}

	results := make([]*knowledge.RetrievalResult, len(chunks))
	for i, chunk := range chunks {
		score := 1.0 / float64(i+1) // Simple rank-based score
		results[i] = &knowledge.RetrievalResult{
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
func (s *Service) GetConfig() knowledge.RetrievalConfig {
	return s.config
}

// RetrieveWithMultiQuery performs retrieval using multiple query variations
// This can improve recall for complex or ambiguous queries
func (s *Service) RetrieveWithMultiQuery(ctx context.Context, input RetrieveInput) (*RetrieveResult, error) {
	// Generate multiple query variations
	queryVariations := s.queryProcessor.GenerateMultipleQueries(input.Query)

	if len(queryVariations) <= 1 {
		// Only one query, use standard retrieval
		return s.Retrieve(ctx, input)
	}

	startTime := time.Now()
	stats := &knowledge.RetrievalStats{}

	// Process the original query for metadata
	processedQuery := s.queryProcessor.Process(input.Query)

	// Collect results from all query variations
	allResultsMap := make(map[common.ID]*knowledge.RetrievalResult)
	queryScores := make(map[common.ID][]float64) // Track scores across queries

	for _, queryVar := range queryVariations {
		varInput := RetrieveInput{
			Query:             queryVar,
			WorkspaceID:       input.WorkspaceID,
			ProjectID:         input.ProjectID,
			MaxResults:        s.config.VectorTopK, // Get more results per query
			EnableQueryExpand: false,               // Already expanded
		}

		result, err := s.Retrieve(ctx, varInput)
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
	results := make([]*knowledge.RetrievalResult, 0, len(allResultsMap))
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
		maxResults = s.config.FinalTopK
	}
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	if input.MaxTokens > 0 {
		results = s.applyTokenLimit(results, input.MaxTokens)
	}

	stats.TotalTimeMs = time.Since(startTime).Milliseconds()

	return &RetrieveResult{
		Results:        results,
		Stats:          stats,
		ProcessedQuery: processedQuery,
	}, nil
}

// GetQueryProcessor returns the query processor for external use
func (s *Service) GetQueryProcessor() *QueryProcessor {
	return s.queryProcessor
}
