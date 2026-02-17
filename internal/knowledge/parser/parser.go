package parser

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
)

// ParsedDocument is the unified output from all parsers.
type ParsedDocument struct {
	Content  string           // Extracted text (markdown-formatted)
	Title    string           // Extracted from document or filename-based
	Metadata DocumentMetadata // Format-specific metadata
}

// DocumentMetadata holds format-specific information about the parsed document.
type DocumentMetadata struct {
	SourceFormat string // "pdf", "docx", "html", "markdown", "text", "code"
	PageCount    int    // For PDF
	WordCount    int
	Author       string // From document metadata if available
	CodeLanguage string // For code files
}

// DocumentParser extracts text content from a specific format.
type DocumentParser interface {
	Parse(ctx context.Context, filename string, reader io.Reader) (*ParsedDocument, error)
	SupportedFormats() []string // MIME types this parser handles
}

// FileValidator validates uploaded files before parsing.
type FileValidator struct {
	MaxSizeBytes     int64
	AllowedMIMETypes map[string]bool
}

// NewFileValidator creates a validator with sensible defaults.
func NewFileValidator() *FileValidator {
	allowed := map[string]bool{
		// Documents
		"application/pdf": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"text/html":     true,
		"text/markdown": true,
		"text/plain":    true,
		// Code files — mapped via extension below
		"text/x-go":         true,
		"text/x-python":     true,
		"text/x-java":       true,
		"text/javascript":   true,
		"text/typescript":   true,
		"text/x-c":          true,
		"text/x-c++":        true,
		"text/x-ruby":       true,
		"text/x-rust":       true,
		"text/x-shellscript": true,
		"text/yaml":         true,
		"application/json":  true,
		"application/xml":   true,
	}

	return &FileValidator{
		MaxSizeBytes:     50 * 1024 * 1024, // 50 MB
		AllowedMIMETypes: allowed,
	}
}

// Validate checks file size and MIME type.
func (v *FileValidator) Validate(filename string, size int64) error {
	if size > v.MaxSizeBytes {
		return fmt.Errorf("file exceeds maximum size of %d MB", v.MaxSizeBytes/(1024*1024))
	}
	if size == 0 {
		return fmt.Errorf("file is empty")
	}

	mimeType := DetectMIMEType(filename)
	if !v.AllowedMIMETypes[mimeType] {
		return fmt.Errorf("unsupported file type: %s (detected MIME: %s)", filepath.Ext(filename), mimeType)
	}
	return nil
}

// Registry maps MIME types to parsers.
type Registry struct {
	parsers map[string]DocumentParser
}

// NewRegistry creates a registry pre-loaded with all built-in parsers.
func NewRegistry() *Registry {
	r := &Registry{parsers: make(map[string]DocumentParser)}

	// Register all built-in parsers
	builtins := []DocumentParser{
		NewPDFParser(),
		NewDOCXParser(),
		NewHTMLParser(),
		NewMarkdownParser(),
		NewTextParser(),
		NewCodeParser(),
	}

	for _, p := range builtins {
		for _, mimeType := range p.SupportedFormats() {
			r.parsers[mimeType] = p
		}
	}

	return r
}

// GetParser returns the parser for a given filename, or an error if unsupported.
func (r *Registry) GetParser(filename string) (DocumentParser, error) {
	mimeType := DetectMIMEType(filename)
	p, ok := r.parsers[mimeType]
	if !ok {
		return nil, fmt.Errorf("no parser for MIME type %q (file: %s)", mimeType, filename)
	}
	return p, nil
}

// Parse is a convenience method that finds the right parser and parses the file.
func (r *Registry) Parse(ctx context.Context, filename string, reader io.Reader) (*ParsedDocument, error) {
	p, err := r.GetParser(filename)
	if err != nil {
		return nil, err
	}
	return p.Parse(ctx, filename, reader)
}

// DetectMIMEType determines MIME type from filename extension.
func DetectMIMEType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	// Custom overrides for common types that mime.TypeByExtension may miss
	overrides := map[string]string{
		".md":    "text/markdown",
		".mdx":   "text/markdown",
		".go":    "text/x-go",
		".py":    "text/x-python",
		".java":  "text/x-java",
		".js":    "text/javascript",
		".ts":    "text/typescript",
		".tsx":   "text/typescript",
		".jsx":   "text/javascript",
		".c":     "text/x-c",
		".cpp":   "text/x-c++",
		".h":     "text/x-c",
		".hpp":   "text/x-c++",
		".rb":    "text/x-ruby",
		".rs":    "text/x-rust",
		".sh":    "text/x-shellscript",
		".bash":  "text/x-shellscript",
		".yaml":  "text/yaml",
		".yml":   "text/yaml",
		".json":  "application/json",
		".xml":   "application/xml",
		".html":  "text/html",
		".htm":   "text/html",
		".pdf":   "application/pdf",
		".docx":  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".txt":   "text/plain",
		".csv":   "text/plain",
		".log":   "text/plain",
		".cfg":   "text/plain",
		".ini":   "text/plain",
		".toml":  "text/plain",
		".env":   "text/plain",
		".sql":   "text/x-sql",
		".proto": "text/x-protobuf",
		".tf":    "text/x-terraform",
	}

	if mt, ok := overrides[ext]; ok {
		return mt
	}

	// Fall back to Go's mime package
	if mt := mime.TypeByExtension(ext); mt != "" {
		// Normalize: strip parameters like charset
		if idx := strings.Index(mt, ";"); idx >= 0 {
			mt = strings.TrimSpace(mt[:idx])
		}
		return mt
	}

	return "application/octet-stream"
}

// countWords counts the number of words in a text.
func countWords(text string) int {
	return len(strings.Fields(text))
}

// titleFromFilename derives a title from the filename (without extension).
func titleFromFilename(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	// Replace common separators with spaces
	name = strings.NewReplacer("-", " ", "_", " ").Replace(name)
	return strings.TrimSpace(name)
}
