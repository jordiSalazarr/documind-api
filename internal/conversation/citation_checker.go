package conversation

import (
	"regexp"
	"strings"

	"documind.jordi.org/internal/conversation/features/commands/sendmessage"
)

// CoverageResult contains the result of a citation coverage check.
type CoverageResult struct {
	CoverageRatio    float64 // Ratio of cited sentences to total sentences (0.0-1.0)
	TotalSentences   int     // Total sentences in the answer
	CitedSentences   int     // Sentences containing at least one [Source N] citation
	UncitedSentences int     // Sentences without any citation
	CitedSources     []int   // Unique source indices referenced in the answer
}

// CitationChecker verifies that LLM responses properly cite their sources.
type CitationChecker struct {
	sourcePattern   *regexp.Regexp
	sentenceSplitRe *regexp.Regexp
}

// NewCitationChecker creates a new citation checker.
func NewCitationChecker() *CitationChecker {
	return &CitationChecker{
		sourcePattern:   regexp.MustCompile(`\[Source\s+(\d+)\]`),
		sentenceSplitRe: regexp.MustCompile(`[.!?]+\s+`),
	}
}

// CheckCoverage analyzes an answer for citation coverage.
func (cc *CitationChecker) CheckCoverage(answer string, sourceCount int) CoverageResult {
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return CoverageResult{}
	}

	sentences := cc.splitSentences(answer)
	if len(sentences) == 0 {
		return CoverageResult{}
	}

	// Count cited vs uncited sentences
	cited := 0
	sourceSet := make(map[int]struct{})

	for _, sentence := range sentences {
		matches := cc.sourcePattern.FindAllStringSubmatch(sentence, -1)
		if len(matches) > 0 {
			cited++
			for _, match := range matches {
				if len(match) > 1 {
					var idx int
					if _, err := parseSourceIndex(match[1]); err == nil {
						idx, _ = parseSourceIndex(match[1])
						sourceSet[idx] = struct{}{}
					}
				}
			}
		}
	}

	uncited := len(sentences) - cited

	// Collect unique source indices
	citedSources := make([]int, 0, len(sourceSet))
	for idx := range sourceSet {
		citedSources = append(citedSources, idx)
	}

	var ratio float64
	if len(sentences) > 0 {
		ratio = float64(cited) / float64(len(sentences))
	}

	return CoverageResult{
		CoverageRatio:    ratio,
		TotalSentences:   len(sentences),
		CitedSentences:   cited,
		UncitedSentences: uncited,
		CitedSources:     citedSources,
	}
}

// LowCoverageDisclaimer is appended when citation coverage falls below threshold.
const LowCoverageDisclaimer = "\n\n---\n*Note: Parts of this answer may not be directly supported by the available documents.*"

// splitSentences splits text into sentences, filtering out very short fragments.
func (cc *CitationChecker) splitSentences(text string) []string {
	// Split on sentence boundaries
	raw := cc.sentenceSplitRe.Split(text, -1)

	sentences := make([]string, 0, len(raw))
	for _, s := range raw {
		s = strings.TrimSpace(s)
		// Skip very short fragments (headers, single words, bullet markers)
		if len(s) > 20 {
			sentences = append(sentences, s)
		}
	}

	// Fallback: if no sentences detected, treat entire text as one sentence
	if len(sentences) == 0 && len(text) > 20 {
		sentences = append(sentences, text)
	}

	return sentences
}

// parseSourceIndex parses a source index string to int.
func parseSourceIndex(s string) (int, error) {
	var idx int
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		idx = idx*10 + int(c-'0')
	}
	return idx, nil
}

// citationCheckerAdapter adapts CitationChecker to the sendmessage.CitationCoverageChecker interface.
type citationCheckerAdapter struct {
	checker *CitationChecker
}

func newCitationCheckerAdapter() *citationCheckerAdapter {
	return &citationCheckerAdapter{checker: NewCitationChecker()}
}

func (a *citationCheckerAdapter) CheckCoverage(answer string, sourceCount int) sendmessage.CitationCoverageResult {
	result := a.checker.CheckCoverage(answer, sourceCount)
	return sendmessage.CitationCoverageResult{
		CoverageRatio:    result.CoverageRatio,
		TotalSentences:   result.TotalSentences,
		CitedSentences:   result.CitedSentences,
		UncitedSentences: result.UncitedSentences,
		CitedSources:     result.CitedSources,
	}
}
