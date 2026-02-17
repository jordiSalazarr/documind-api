package parser

import (
	"context"
	"io"
	"regexp"
	"strings"
)

// MarkdownParser is a passthrough parser for markdown files.
// The content is already in the format the chunker expects.
type MarkdownParser struct {
	titleRe *regexp.Regexp
}

// NewMarkdownParser creates a new markdown parser.
func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{
		titleRe: regexp.MustCompile(`(?m)^#\s+(.+)$`),
	}
}

// SupportedFormats returns the MIME types this parser handles.
func (p *MarkdownParser) SupportedFormats() []string {
	return []string{"text/markdown"}
}

// Parse reads markdown content and extracts a title from the first H1 heading.
func (p *MarkdownParser) Parse(ctx context.Context, filename string, reader io.Reader) (*ParsedDocument, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	content := string(data)

	// Try to extract title from first H1 heading
	title := ""
	if match := p.titleRe.FindStringSubmatch(content); len(match) > 1 {
		title = strings.TrimSpace(match[1])
	}
	if title == "" {
		title = titleFromFilename(filename)
	}

	return &ParsedDocument{
		Content: content,
		Title:   title,
		Metadata: DocumentMetadata{
			SourceFormat: "markdown",
			WordCount:    countWords(content),
		},
	}, nil
}
