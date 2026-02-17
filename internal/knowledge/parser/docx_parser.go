package parser

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/nguyenthenguyen/docx"
)

// DOCXParser extracts text content from DOCX files.
type DOCXParser struct {
	tagRe *regexp.Regexp
}

// NewDOCXParser creates a new DOCX parser.
func NewDOCXParser() *DOCXParser {
	return &DOCXParser{
		tagRe: regexp.MustCompile(`<[^>]+>`),
	}
}

// SupportedFormats returns the MIME types this parser handles.
func (p *DOCXParser) SupportedFormats() []string {
	return []string{"application/vnd.openxmlformats-officedocument.wordprocessingml.document"}
}

// Parse extracts text from a DOCX file.
func (p *DOCXParser) Parse(ctx context.Context, filename string, reader io.Reader) (*ParsedDocument, error) {
	// The docx library requires a file path, so we write to a temp file
	tmpFile, err := os.CreateTemp("", "documind-docx-*.docx")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, reader); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	doc, err := docx.ReadDocxFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to open DOCX: %w", err)
	}
	defer doc.Close()

	editable := doc.Editable()
	rawContent := editable.GetContent()

	// Convert DOCX XML content to plain text / markdown
	content := p.convertToMarkdown(rawContent)

	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("no extractable text found in DOCX")
	}

	title := titleFromFilename(filename)

	return &ParsedDocument{
		Content: content,
		Title:   title,
		Metadata: DocumentMetadata{
			SourceFormat: "docx",
			WordCount:    countWords(content),
		},
	}, nil
}

// convertToMarkdown converts raw DOCX XML content to readable markdown text.
func (p *DOCXParser) convertToMarkdown(content string) string {
	// Replace common XML paragraph markers with newlines
	content = strings.ReplaceAll(content, "</w:p>", "\n\n")
	content = strings.ReplaceAll(content, "</w:r>", "")
	content = strings.ReplaceAll(content, "<w:br/>", "\n")
	content = strings.ReplaceAll(content, "<w:tab/>", "\t")

	// Strip all remaining XML tags
	content = p.tagRe.ReplaceAllString(content, "")

	// Clean up whitespace
	lines := strings.Split(content, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}

	return strings.Join(cleaned, "\n\n")
}
