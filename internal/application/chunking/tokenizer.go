package chunking

import (
	"strings"
	"unicode"
)

// Tokenizer provides token counting for text
type Tokenizer interface {
	CountTokens(text string) int
}

// SimpleTokenizer provides a simple token counting implementation
// This is an approximation - for production, use tiktoken-go or similar
type SimpleTokenizer struct{}

// NewSimpleTokenizer creates a new simple tokenizer
func NewSimpleTokenizer() *SimpleTokenizer {
	return &SimpleTokenizer{}
}

// CountTokens estimates token count
// Approximate ratio: 1 token ≈ 4 characters for English text
// This matches OpenAI's estimation for GPT models
func (t *SimpleTokenizer) CountTokens(text string) int {
	if text == "" {
		return 0
	}

	// More accurate estimation using word count + punctuation
	words := 0
	inWord := false

	for _, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			if inWord {
				words++
				inWord = false
			}
		} else {
			inWord = true
		}
	}

	if inWord {
		words++
	}

	// Average token to word ratio for English: ~1.3 tokens per word
	// This accounts for subword tokenization
	tokens := int(float64(words) * 1.3)

	// Ensure minimum of 1 token for non-empty text
	if tokens == 0 && len(text) > 0 {
		return 1
	}

	return tokens
}

// EstimateTokensFromChars provides a quick character-based estimation
func EstimateTokensFromChars(charCount int) int {
	// 1 token ≈ 4 characters
	tokens := charCount / 4
	if tokens == 0 && charCount > 0 {
		return 1
	}
	return tokens
}

// SplitIntoSentences splits text into sentences
func SplitIntoSentences(text string) []string {
	// Simple sentence splitter
	// For production, consider using a proper NLP library
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{}
	}

	sentences := []string{}
	current := strings.Builder{}

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		current.WriteRune(runes[i])

		// Check for sentence boundaries
		if runes[i] == '.' || runes[i] == '!' || runes[i] == '?' {
			// Look ahead to ensure it's actually end of sentence
			if i+1 < len(runes) {
				next := runes[i+1]
				// If followed by space and uppercase, it's likely end of sentence
				if unicode.IsSpace(next) {
					if i+2 < len(runes) && unicode.IsUpper(runes[i+2]) {
						sentence := strings.TrimSpace(current.String())
						if sentence != "" {
							sentences = append(sentences, sentence)
						}
						current.Reset()
					}
				}
			} else {
				// End of text
				sentence := strings.TrimSpace(current.String())
				if sentence != "" {
					sentences = append(sentences, sentence)
				}
				current.Reset()
			}
		}
	}

	// Add remaining text as a sentence
	if current.Len() > 0 {
		sentence := strings.TrimSpace(current.String())
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
	}

	// Fallback: if no sentences detected, treat entire text as one sentence
	if len(sentences) == 0 && text != "" {
		sentences = append(sentences, text)
	}

	return sentences
}

// SplitIntoParagraphs splits text into paragraphs
func SplitIntoParagraphs(text string) []string {
	paragraphs := []string{}

	// Split by double newlines (standard paragraph separator)
	parts := strings.Split(text, "\n\n")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			paragraphs = append(paragraphs, part)
		}
	}

	// Fallback: if no double newlines, split by single newlines
	if len(paragraphs) == 0 {
		parts = strings.Split(text, "\n")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				paragraphs = append(paragraphs, part)
			}
		}
	}

	// Last fallback: treat entire text as one paragraph
	if len(paragraphs) == 0 && text != "" {
		paragraphs = append(paragraphs, strings.TrimSpace(text))
	}

	return paragraphs
}
