package chunker

import (
	"log"
	"strings"
	"unicode"

	tiktoken "github.com/pkoukk/tiktoken-go"
)

// Tokenizer provides token counting for text
type Tokenizer interface {
	CountTokens(text string) int
}

// TiktokenTokenizer provides exact token counting using OpenAI's cl100k_base encoding.
// This matches the tokenizer used by text-embedding-3-small and GPT-4 models.
type TiktokenTokenizer struct {
	encoder *tiktoken.Tiktoken
}

// NewTiktokenTokenizer creates a tokenizer with exact cl100k_base token counts.
// Falls back to SimpleTokenizer if initialization fails.
func NewTiktokenTokenizer() Tokenizer {
	enc, err := tiktoken.EncodingForModel("text-embedding-3-small")
	if err != nil {
		log.Printf("WARNING: tiktoken init failed, falling back to SimpleTokenizer: %v", err)
		return NewSimpleTokenizer()
	}
	return &TiktokenTokenizer{encoder: enc}
}

// CountTokens returns the exact token count using cl100k_base encoding.
func (t *TiktokenTokenizer) CountTokens(text string) int {
	if text == "" {
		return 0
	}
	return len(t.encoder.Encode(text, nil, nil))
}

// SimpleTokenizer provides a simple token counting implementation
// Used as a fallback when tiktoken initialization fails.
type SimpleTokenizer struct{}

// NewSimpleTokenizer creates a new simple tokenizer
func NewSimpleTokenizer() *SimpleTokenizer {
	return &SimpleTokenizer{}
}

// CountTokens estimates token count
// Approximate ratio: 1 token ~ 4 characters for English text
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
	// 1 token ~ 4 characters
	tokens := charCount / 4
	if tokens == 0 && charCount > 0 {
		return 1
	}
	return tokens
}

// SplitIntoSentences splits text into sentences
func SplitIntoSentences(text string) []string {
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
