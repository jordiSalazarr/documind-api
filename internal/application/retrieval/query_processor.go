package retrieval

import (
	"regexp"
	"strings"
	"unicode"
)

// QueryProcessor handles query preprocessing, expansion, and normalization
type QueryProcessor struct {
	acronyms         map[string]string
	synonyms         map[string][]string
	stopWords        map[string]bool
	enableExpansion  bool
	enableCorrection bool
	maxExpansions    int
}

// QueryProcessorConfig holds configuration for query processing
type QueryProcessorConfig struct {
	EnableExpansion  bool // Enable synonym/acronym expansion
	EnableCorrection bool // Enable basic spell correction
	MaxExpansions    int  // Max number of expansion terms to add
	CustomAcronyms   map[string]string
	CustomSynonyms   map[string][]string
}

// DefaultQueryProcessorConfig returns sensible defaults
func DefaultQueryProcessorConfig() QueryProcessorConfig {
	return QueryProcessorConfig{
		EnableExpansion:  true,
		EnableCorrection: true,
		MaxExpansions:    5,
	}
}

// NewQueryProcessor creates a new query processor
func NewQueryProcessor(config QueryProcessorConfig) *QueryProcessor {
	qp := &QueryProcessor{
		acronyms:         buildDefaultAcronyms(),
		synonyms:         buildDefaultSynonyms(),
		stopWords:        buildStopWords(),
		enableExpansion:  config.EnableExpansion,
		enableCorrection: config.EnableCorrection,
		maxExpansions:    config.MaxExpansions,
	}

	if config.MaxExpansions <= 0 {
		qp.maxExpansions = 5
	}

	// Merge custom acronyms
	for k, v := range config.CustomAcronyms {
		qp.acronyms[strings.ToLower(k)] = v
	}

	// Merge custom synonyms
	for k, v := range config.CustomSynonyms {
		qp.synonyms[strings.ToLower(k)] = v
	}

	return qp
}

// ProcessedQuery contains the result of query processing
type ProcessedQuery struct {
	Original        string   // Original query
	Normalized      string   // Cleaned and normalized query
	Expanded        string   // Query with expansions added
	Keywords        []string // Extracted keywords
	Expansions      []string // Terms added via expansion
	Intent          QueryIntent
	IsQuestion      bool
	HasCodeTerms    bool
	SuggestedFields []string // Fields to prioritize in search
}

// QueryIntent represents the detected intent of a query
type QueryIntent int

const (
	IntentUnknown QueryIntent = iota
	IntentFactual             // "What is X?"
	IntentProcedural          // "How to do Y?"
	IntentComparison          // "X vs Y"
	IntentTroubleshooting     // "Error: ...", "fix", "issue"
	IntentExploration         // General exploration
	IntentDefinition          // "Define X", "meaning of"
)

// Process performs full query processing pipeline
func (qp *QueryProcessor) Process(query string) *ProcessedQuery {
	result := &ProcessedQuery{
		Original:   query,
		Keywords:   []string{},
		Expansions: []string{},
	}

	// Step 1: Normalize
	result.Normalized = qp.normalize(query)

	// Step 2: Detect intent
	result.Intent = qp.detectIntent(result.Normalized)
	result.IsQuestion = qp.isQuestion(result.Normalized)

	// Step 3: Extract keywords
	result.Keywords = qp.extractKeywords(result.Normalized)

	// Step 4: Check for code-related terms
	result.HasCodeTerms = qp.hasCodeTerms(result.Normalized)

	// Step 5: Expand query if enabled
	if qp.enableExpansion {
		result.Expanded, result.Expansions = qp.expand(result.Normalized, result.Keywords)
	} else {
		result.Expanded = result.Normalized
	}

	// Step 6: Suggest fields to prioritize
	result.SuggestedFields = qp.suggestFields(result)

	return result
}

// normalize cleans and normalizes the query
func (qp *QueryProcessor) normalize(query string) string {
	// Trim whitespace
	query = strings.TrimSpace(query)

	// Collapse multiple spaces
	spaceRegex := regexp.MustCompile(`\s+`)
	query = spaceRegex.ReplaceAllString(query, " ")

	// Remove special characters but keep useful punctuation
	var normalized strings.Builder
	for _, r := range query {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) ||
			r == '-' || r == '_' || r == '.' || r == '?' || r == ':' || r == '/' {
			normalized.WriteRune(r)
		}
	}

	return normalized.String()
}

// detectIntent classifies the query intent
func (qp *QueryProcessor) detectIntent(query string) QueryIntent {
	lower := strings.ToLower(query)

	// Troubleshooting patterns
	troubleshootingPatterns := []string{
		"error", "issue", "problem", "bug", "fix", "broken",
		"doesn't work", "not working", "failed", "failing",
		"exception", "crash", "debug",
	}
	for _, pattern := range troubleshootingPatterns {
		if strings.Contains(lower, pattern) {
			return IntentTroubleshooting
		}
	}

	// Procedural patterns
	proceduralPatterns := []string{
		"how to", "how do", "how can", "steps to", "guide",
		"tutorial", "setup", "configure", "install", "create",
		"implement", "build", "deploy",
	}
	for _, pattern := range proceduralPatterns {
		if strings.Contains(lower, pattern) {
			return IntentProcedural
		}
	}

	// Comparison patterns
	comparisonPatterns := []string{
		" vs ", " versus ", "compare", "difference between",
		"better than", "comparison", " or ",
	}
	for _, pattern := range comparisonPatterns {
		if strings.Contains(lower, pattern) {
			return IntentComparison
		}
	}

	// Definition patterns
	definitionPatterns := []string{
		"what is", "what are", "define", "meaning of",
		"definition", "explain what",
	}
	for _, pattern := range definitionPatterns {
		if strings.Contains(lower, pattern) {
			return IntentDefinition
		}
	}

	// Factual patterns
	factualPatterns := []string{
		"what", "when", "where", "who", "which",
	}
	for _, pattern := range factualPatterns {
		if strings.HasPrefix(lower, pattern) {
			return IntentFactual
		}
	}

	return IntentExploration
}

// isQuestion determines if the query is a question
func (qp *QueryProcessor) isQuestion(query string) bool {
	lower := strings.ToLower(query)

	// Check for question mark
	if strings.HasSuffix(query, "?") {
		return true
	}

	// Check for question words at start
	questionStarters := []string{
		"what", "how", "why", "when", "where", "who", "which",
		"can", "could", "would", "should", "is", "are", "do", "does",
	}
	for _, starter := range questionStarters {
		if strings.HasPrefix(lower, starter+" ") {
			return true
		}
	}

	return false
}

// extractKeywords extracts meaningful keywords from query
func (qp *QueryProcessor) extractKeywords(query string) []string {
	words := strings.Fields(strings.ToLower(query))
	keywords := make([]string, 0, len(words))

	for _, word := range words {
		// Clean punctuation
		word = strings.Trim(word, ".,?!:;\"'")

		// Skip stop words and short words
		if len(word) < 2 {
			continue
		}
		if qp.stopWords[word] {
			continue
		}

		keywords = append(keywords, word)
	}

	return keywords
}

// hasCodeTerms checks if query contains programming-related terms
func (qp *QueryProcessor) hasCodeTerms(query string) bool {
	lower := strings.ToLower(query)

	codeTerms := []string{
		"function", "method", "class", "variable", "const",
		"import", "export", "async", "await", "return",
		"api", "endpoint", "request", "response", "json",
		"database", "query", "sql", "table", "schema",
		"error", "exception", "debug", "log", "trace",
		"git", "commit", "branch", "merge", "pull",
		"docker", "container", "kubernetes", "deploy",
		"test", "unit", "integration", "mock",
		"()", "{}", "[]", "=>", "->",
	}

	for _, term := range codeTerms {
		if strings.Contains(lower, term) {
			return true
		}
	}

	return false
}

// expand adds synonyms and acronym expansions to the query
func (qp *QueryProcessor) expand(query string, keywords []string) (string, []string) {
	expansions := []string{}
	expandedTerms := make(map[string]bool)

	for _, keyword := range keywords {
		lower := strings.ToLower(keyword)

		// Check acronyms
		if expansion, ok := qp.acronyms[lower]; ok {
			if !expandedTerms[expansion] {
				expansions = append(expansions, expansion)
				expandedTerms[expansion] = true
			}
		}

		// Check synonyms
		if syns, ok := qp.synonyms[lower]; ok {
			for _, syn := range syns {
				if !expandedTerms[syn] && len(expansions) < qp.maxExpansions {
					expansions = append(expansions, syn)
					expandedTerms[syn] = true
				}
			}
		}
	}

	// Build expanded query
	if len(expansions) > 0 {
		return query + " " + strings.Join(expansions, " "), expansions
	}
	return query, expansions
}

// suggestFields suggests which fields to prioritize based on query
func (qp *QueryProcessor) suggestFields(pq *ProcessedQuery) []string {
	fields := []string{}

	switch pq.Intent {
	case IntentProcedural:
		fields = append(fields, "content", "heading")
	case IntentTroubleshooting:
		fields = append(fields, "content", "metadata.code_block")
	case IntentDefinition:
		fields = append(fields, "heading", "content")
	case IntentComparison:
		fields = append(fields, "content", "heading")
	default:
		fields = append(fields, "content", "heading")
	}

	if pq.HasCodeTerms {
		fields = append(fields, "metadata.code_block", "metadata.language")
	}

	return fields
}

// GenerateMultipleQueries generates query variations for better recall
func (qp *QueryProcessor) GenerateMultipleQueries(query string) []string {
	queries := []string{query}
	processed := qp.Process(query)

	// Add expanded version
	if processed.Expanded != query && processed.Expanded != "" {
		queries = append(queries, processed.Expanded)
	}

	// Generate keyword-focused version (remove question words)
	keywordQuery := strings.Join(processed.Keywords, " ")
	if keywordQuery != "" && keywordQuery != query {
		queries = append(queries, keywordQuery)
	}

	// For procedural queries, add action-focused version
	if processed.Intent == IntentProcedural {
		// Extract the action verb + object
		actionQuery := qp.extractActionPhrase(query)
		if actionQuery != "" && actionQuery != query {
			queries = append(queries, actionQuery)
		}
	}

	return queries
}

// extractActionPhrase extracts the main action from a procedural query
func (qp *QueryProcessor) extractActionPhrase(query string) string {
	lower := strings.ToLower(query)

	// Remove common prefixes
	prefixes := []string{
		"how to ", "how do i ", "how can i ", "how do you ",
		"steps to ", "guide to ", "tutorial for ",
		"i want to ", "i need to ", "help me ",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(lower, prefix) {
			return strings.TrimPrefix(lower, prefix)
		}
	}

	return ""
}

// buildDefaultAcronyms returns common tech acronyms
func buildDefaultAcronyms() map[string]string {
	return map[string]string{
		"api":   "application programming interface",
		"jwt":   "json web token",
		"auth":  "authentication authorization",
		"db":    "database",
		"sql":   "structured query language",
		"nosql": "non-relational database",
		"ui":    "user interface",
		"ux":    "user experience",
		"css":   "cascading style sheets",
		"html":  "hypertext markup language",
		"http":  "hypertext transfer protocol",
		"https": "secure http",
		"rest":  "representational state transfer",
		"crud":  "create read update delete",
		"orm":   "object relational mapping",
		"cli":   "command line interface",
		"sdk":   "software development kit",
		"ide":   "integrated development environment",
		"ci":    "continuous integration",
		"cd":    "continuous deployment",
		"devops": "development operations",
		"k8s":   "kubernetes",
		"aws":   "amazon web services",
		"gcp":   "google cloud platform",
		"vm":    "virtual machine",
		"os":    "operating system",
		"env":   "environment",
		"config": "configuration",
		"repo":  "repository",
		"pr":    "pull request",
		"mr":    "merge request",
		"rag":   "retrieval augmented generation",
		"llm":   "large language model",
		"ml":    "machine learning",
		"ai":    "artificial intelligence",
		"nlp":   "natural language processing",
	}
}

// buildDefaultSynonyms returns common synonym mappings
func buildDefaultSynonyms() map[string][]string {
	return map[string][]string{
		"create":     {"make", "build", "generate", "new"},
		"delete":     {"remove", "destroy", "drop"},
		"update":     {"modify", "change", "edit", "patch"},
		"get":        {"fetch", "retrieve", "read", "obtain"},
		"error":      {"exception", "issue", "problem", "bug"},
		"fix":        {"resolve", "repair", "solve", "debug"},
		"install":    {"setup", "configure", "deploy"},
		"start":      {"begin", "launch", "run", "initialize"},
		"stop":       {"end", "terminate", "halt", "shutdown"},
		"user":       {"account", "member", "person"},
		"password":   {"credential", "secret", "passphrase"},
		"login":      {"signin", "authenticate", "logon"},
		"logout":     {"signout", "logoff"},
		"settings":   {"configuration", "preferences", "options"},
		"fast":       {"quick", "rapid", "performance"},
		"slow":       {"lag", "delay", "latency"},
		"file":       {"document", "asset"},
		"folder":     {"directory", "path"},
		"search":     {"find", "query", "lookup"},
		"test":       {"verify", "validate", "check"},
	}
}

// buildStopWords returns common English stop words
func buildStopWords() map[string]bool {
	words := []string{
		"a", "an", "the", "and", "or", "but", "in", "on", "at", "to", "for",
		"of", "with", "by", "from", "as", "is", "was", "are", "were", "been",
		"be", "have", "has", "had", "do", "does", "did", "will", "would",
		"could", "should", "may", "might", "must", "shall", "can", "need",
		"this", "that", "these", "those", "i", "you", "he", "she", "it",
		"we", "they", "what", "which", "who", "whom", "when", "where", "why",
		"how", "all", "each", "every", "both", "few", "more", "most", "other",
		"some", "such", "no", "nor", "not", "only", "own", "same", "so",
		"than", "too", "very", "just", "also", "now", "here", "there",
	}

	stopWords := make(map[string]bool, len(words))
	for _, w := range words {
		stopWords[w] = true
	}
	return stopWords
}
