package parser

import (
	"context"
	"io"
)

// TextParser is a passthrough parser for plain text files.
type TextParser struct{}

// NewTextParser creates a new text parser.
func NewTextParser() *TextParser {
	return &TextParser{}
}

// SupportedFormats returns the MIME types this parser handles.
func (p *TextParser) SupportedFormats() []string {
	return []string{"text/plain"}
}

// Parse reads plain text content.
func (p *TextParser) Parse(ctx context.Context, filename string, reader io.Reader) (*ParsedDocument, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	content := string(data)

	return &ParsedDocument{
		Content: content,
		Title:   titleFromFilename(filename),
		Metadata: DocumentMetadata{
			SourceFormat: "text",
			WordCount:    countWords(content),
		},
	}, nil
}
