package parser

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	readability "github.com/go-shiori/go-readability"
)

// HTMLParser extracts main content from HTML files using readability.
type HTMLParser struct{}

// NewHTMLParser creates a new HTML parser.
func NewHTMLParser() *HTMLParser {
	return &HTMLParser{}
}

// SupportedFormats returns the MIME types this parser handles.
func (p *HTMLParser) SupportedFormats() []string {
	return []string{"text/html"}
}

// Parse extracts main content from an HTML file.
func (p *HTMLParser) Parse(ctx context.Context, filename string, reader io.Reader) (*ParsedDocument, error) {
	// Use go-readability to extract main content (strips nav, footer, ads, etc.)
	fakeURL := &url.URL{Scheme: "file", Path: "/" + filename}
	article, err := readability.FromReader(reader, fakeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// TextContent is plain text; Content is cleaned HTML.
	// We use TextContent for the markdown output.
	content := strings.TrimSpace(article.TextContent)
	if content == "" {
		return nil, fmt.Errorf("no extractable text found in HTML")
	}

	// Use the article title if available, otherwise fall back to filename
	title := article.Title
	if title == "" {
		title = titleFromFilename(filename)
	}

	return &ParsedDocument{
		Content: content,
		Title:   title,
		Metadata: DocumentMetadata{
			SourceFormat: "html",
			WordCount:    countWords(content),
			Author:       article.Byline,
		},
	}, nil
}
