package chunker

import (
	"fmt"
	"regexp"
	"strings"

	shared "documind.jordi.org/internal/shared/domain"
	chunkdomain "documind.jordi.org/internal/retrieval/domain"
)

// ChunkingConfig holds configuration for the semantic chunker
type ChunkingConfig struct {
	MaxTokens     int // Maximum tokens per chunk (default: 512)
	TargetTokens  int // Target tokens per chunk (default: 384)
	OverlapTokens int // Overlap between chunks (default: 64)
	MinTokens     int // Minimum tokens per chunk (default: 50)
}

// DefaultChunkingConfig returns the default configuration
func DefaultChunkingConfig() ChunkingConfig {
	return ChunkingConfig{
		MaxTokens:     512,
		TargetTokens:  384,
		OverlapTokens: 64,
		MinTokens:     50,
	}
}

// SemanticChunker implements intelligent document chunking
type SemanticChunker struct {
	config    ChunkingConfig
	tokenizer Tokenizer
}

// NewSemanticChunker creates a new semantic chunker with accurate tiktoken tokenizer
func NewSemanticChunker(config ChunkingConfig) *SemanticChunker {
	return &SemanticChunker{
		config:    config,
		tokenizer: NewTiktokenTokenizer(),
	}
}

// MarkdownSection represents a section in a markdown document
type MarkdownSection struct {
	Heading          string
	HeadingHierarchy []string // Full heading path (e.g., ["Auth Guide", "JWT Setup", "Configuration"])
	Content          string
	Level            int // Header level (1-6)
}

// ChunkDocument chunks a document into semantically coherent pieces
func (c *SemanticChunker) ChunkDocument(
	itemVersionID shared.ID,
	workspaceID shared.ID,
	title string,
	summary string,
	bodyMd string,
) ([]*chunkdomain.Chunk, error) {
	// Combine title, summary, and body for chunking
	fullText := c.buildFullText(title, summary, bodyMd)

	// Parse markdown structure for intelligent splitting
	sections := c.parseMarkdownSections(bodyMd)

	chunks := []*chunkdomain.Chunk{}
	chunkIndex := 0
	charOffset := 0

	// If we have clear sections, chunk by section
	if len(sections) > 0 {
		for _, section := range sections {
			sectionChunks := c.chunkSection(
				itemVersionID,
				workspaceID,
				section,
				&chunkIndex,
				&charOffset,
			)
			chunks = append(chunks, sectionChunks...)
		}
	} else {
		// No clear sections, use paragraph-based chunking
		sectionChunks := c.chunkSection(
			itemVersionID,
			workspaceID,
			MarkdownSection{
				Heading: title,
				Content: fullText,
				Level:   0,
			},
			&chunkIndex,
			&charOffset,
		)
		chunks = append(chunks, sectionChunks...)
	}

	// Add context prefix with heading hierarchy to each chunk
	c.addContextPrefix(chunks, title)

	return chunks, nil
}

// parseMarkdownSections extracts sections from markdown, tracking heading hierarchy
func (c *SemanticChunker) parseMarkdownSections(markdown string) []MarkdownSection {
	sections := []MarkdownSection{}

	// Regex to match markdown headers (# Header, ## Header, etc.)
	headerRegex := regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)

	lines := strings.Split(markdown, "\n")

	// headingStack tracks the current heading at each level (index 0 = h1, 1 = h2, etc.)
	headingStack := make([]string, 7) // levels 1-6, index 0 unused

	currentSection := MarkdownSection{
		Level:   0,
		Content: "",
	}

	for i, line := range lines {
		match := headerRegex.FindStringSubmatch(line)

		if match != nil {
			// Found a header - save previous section if it has content
			if currentSection.Content != "" || currentSection.Heading != "" {
				sections = append(sections, currentSection)
			}

			// Start new section
			level := len(match[1]) // Count # characters
			heading := match[2]

			// Update heading stack: set this level and clear deeper levels
			headingStack[level] = heading
			for j := level + 1; j <= 6; j++ {
				headingStack[j] = ""
			}

			// Build hierarchy from stack (all non-empty headings up to current level)
			hierarchy := make([]string, 0, level)
			for j := 1; j <= level; j++ {
				if headingStack[j] != "" {
					hierarchy = append(hierarchy, headingStack[j])
				}
			}

			currentSection = MarkdownSection{
				Heading:          heading,
				HeadingHierarchy: hierarchy,
				Level:            level,
				Content:          "",
			}
		} else {
			// Add content to current section
			if i > 0 || currentSection.Heading != "" {
				if currentSection.Content != "" {
					currentSection.Content += "\n"
				}
				currentSection.Content += line
			}
		}
	}

	// Add final section
	if currentSection.Content != "" || currentSection.Heading != "" {
		sections = append(sections, currentSection)
	}

	return sections
}

// chunkSection chunks a single section
func (c *SemanticChunker) chunkSection(
	itemVersionID shared.ID,
	workspaceID shared.ID,
	section MarkdownSection,
	chunkIndex *int,
	charOffset *int,
) []*chunkdomain.Chunk {
	chunks := []*chunkdomain.Chunk{}

	content := section.Content
	if content == "" {
		return chunks
	}

	// Build the heading string from hierarchy (e.g., "Auth Guide > JWT Setup > Config")
	heading := c.buildHeadingPath(section)

	// Check if entire section fits in one chunk
	tokenCount := c.tokenizer.CountTokens(content)
	if tokenCount <= c.config.MaxTokens {
		chunk := c.createChunk(
			itemVersionID,
			workspaceID,
			content,
			*chunkIndex,
			chunkdomain.ChunkLevelSection,
			heading,
			*charOffset,
		)
		chunks = append(chunks, chunk)
		*chunkIndex++
		*charOffset += len(content)
		return chunks
	}

	// Section is too large - split by paragraphs
	paragraphs := SplitIntoParagraphs(content)
	currentChunkContent := strings.Builder{}
	currentChunkTokens := 0
	startOffset := *charOffset

	for i, para := range paragraphs {
		paraTokens := c.tokenizer.CountTokens(para)

		// Check if adding this paragraph would exceed max tokens
		if currentChunkTokens+paraTokens > c.config.MaxTokens && currentChunkContent.Len() > 0 {
			// Save current chunk
			chunkContent := currentChunkContent.String()
			chunk := c.createChunk(
				itemVersionID,
				workspaceID,
				chunkContent,
				*chunkIndex,
				chunkdomain.ChunkLevelParagraph,
				heading,
				startOffset,
			)
			chunks = append(chunks, chunk)
			*chunkIndex++

			// Start new chunk with overlap
			overlapContent := c.getOverlapContent(paragraphs, i, c.config.OverlapTokens)
			currentChunkContent.Reset()
			currentChunkContent.WriteString(overlapContent)
			currentChunkTokens = c.tokenizer.CountTokens(overlapContent)
			startOffset = *charOffset
		}

		// Add paragraph to current chunk
		if currentChunkContent.Len() > 0 {
			currentChunkContent.WriteString("\n\n")
		}
		currentChunkContent.WriteString(para)
		currentChunkTokens += paraTokens
		*charOffset += len(para) + 2 // +2 for \n\n
	}

	// Add final chunk if it has content
	if currentChunkContent.Len() > 0 {
		chunkContent := currentChunkContent.String()
		chunk := c.createChunk(
			itemVersionID,
			workspaceID,
			chunkContent,
			*chunkIndex,
			chunkdomain.ChunkLevelParagraph,
			heading,
			startOffset,
		)
		chunks = append(chunks, chunk)
		*chunkIndex++
	}

	return chunks
}

// buildHeadingPath builds a " > "-separated heading path from the section's hierarchy
func (c *SemanticChunker) buildHeadingPath(section MarkdownSection) string {
	if len(section.HeadingHierarchy) > 0 {
		return strings.Join(section.HeadingHierarchy, " > ")
	}
	return section.Heading
}

// createChunk creates a new chunk with metadata
func (c *SemanticChunker) createChunk(
	itemVersionID shared.ID,
	workspaceID shared.ID,
	content string,
	chunkIndex int,
	level chunkdomain.ChunkLevel,
	heading string,
	startOffset int,
) *chunkdomain.Chunk {
	tokenCount := c.tokenizer.CountTokens(content)

	chunk := chunkdomain.NewChunk(
		itemVersionID,
		workspaceID,
		content,
		chunkIndex,
		level,
		tokenCount,
	)

	if heading != "" {
		chunk.SetHeading(heading)
	}

	chunk.SetCharOffsets(startOffset, startOffset+len(content))

	return chunk
}

// getOverlapContent gets the last N tokens worth of content for overlap
func (c *SemanticChunker) getOverlapContent(paragraphs []string, currentIndex int, overlapTokens int) string {
	if overlapTokens == 0 || currentIndex == 0 {
		return ""
	}

	overlap := strings.Builder{}
	tokensCollected := 0

	// Walk backwards through paragraphs to collect overlap
	for i := currentIndex - 1; i >= 0 && tokensCollected < overlapTokens; i-- {
		para := paragraphs[i]
		paraTokens := c.tokenizer.CountTokens(para)

		if tokensCollected+paraTokens > overlapTokens {
			// Take only part of this paragraph (last sentences)
			sentences := SplitIntoSentences(para)
			for j := len(sentences) - 1; j >= 0 && tokensCollected < overlapTokens; j-- {
				sentTokens := c.tokenizer.CountTokens(sentences[j])
				if tokensCollected+sentTokens <= overlapTokens {
					if overlap.Len() > 0 {
						overlap.WriteString(" ")
					}
					// Prepend sentence
					tempOverlap := overlap.String()
					overlap.Reset()
					overlap.WriteString(sentences[j])
					if tempOverlap != "" {
						overlap.WriteString(" ")
						overlap.WriteString(tempOverlap)
					}
					tokensCollected += sentTokens
				}
			}
			break
		}

		// Add entire paragraph to overlap
		if overlap.Len() > 0 {
			// Prepend paragraph
			tempOverlap := overlap.String()
			overlap.Reset()
			overlap.WriteString(para)
			overlap.WriteString("\n\n")
			overlap.WriteString(tempOverlap)
		} else {
			overlap.WriteString(para)
		}
		tokensCollected += paraTokens
	}

	return overlap.String()
}

// buildFullText combines title, summary, and body
func (c *SemanticChunker) buildFullText(title, summary, body string) string {
	parts := []string{}

	if title != "" {
		parts = append(parts, "# "+title)
	}
	if summary != "" {
		parts = append(parts, summary)
	}
	if body != "" {
		parts = append(parts, body)
	}

	return strings.Join(parts, "\n\n")
}

// addContextPrefix prepends "[Title > Section > Subsection] " to each chunk's content.
// If the chunk already has a heading hierarchy, the title is prepended to it.
func (c *SemanticChunker) addContextPrefix(chunks []*chunkdomain.Chunk, title string) {
	if title == "" {
		return
	}

	for _, chunk := range chunks {
		if strings.HasPrefix(chunk.Content, title) {
			continue
		}

		// Build prefix: [Title > Heading Hierarchy] or [Title] if no heading
		var prefix string
		if chunk.Heading != "" && chunk.Heading != title {
			prefix = fmt.Sprintf("[%s > %s] ", title, chunk.Heading)
		} else {
			prefix = fmt.Sprintf("[%s] ", title)
		}

		chunk.Content = prefix + chunk.Content
		chunk.CharCount = len(chunk.Content)
		chunk.TokenCount = c.tokenizer.CountTokens(chunk.Content)
	}
}
