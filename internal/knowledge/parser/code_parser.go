package parser

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// CodeParser wraps code files in language-tagged markdown code blocks.
type CodeParser struct{}

// NewCodeParser creates a new code parser.
func NewCodeParser() *CodeParser {
	return &CodeParser{}
}

// SupportedFormats returns the MIME types this parser handles.
func (p *CodeParser) SupportedFormats() []string {
	return []string{
		"text/x-go",
		"text/x-python",
		"text/x-java",
		"text/javascript",
		"text/typescript",
		"text/x-c",
		"text/x-c++",
		"text/x-ruby",
		"text/x-rust",
		"text/x-shellscript",
		"text/yaml",
		"application/json",
		"application/xml",
		"text/x-sql",
		"text/x-protobuf",
		"text/x-terraform",
	}
}

// Parse reads a code file and wraps it in a markdown code block.
func (p *CodeParser) Parse(ctx context.Context, filename string, reader io.Reader) (*ParsedDocument, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	raw := string(data)
	lang := detectCodeLanguage(filename)

	content := fmt.Sprintf("```%s\n%s\n```", lang, raw)

	return &ParsedDocument{
		Content: content,
		Title:   titleFromFilename(filename),
		Metadata: DocumentMetadata{
			SourceFormat: "code",
			WordCount:    countWords(raw),
			CodeLanguage: lang,
		},
	}, nil
}

// detectCodeLanguage maps file extension to a markdown code-fence language tag.
func detectCodeLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	langs := map[string]string{
		".go":    "go",
		".py":    "python",
		".java":  "java",
		".js":    "javascript",
		".jsx":   "javascript",
		".ts":    "typescript",
		".tsx":   "typescript",
		".c":     "c",
		".cpp":   "cpp",
		".h":     "c",
		".hpp":   "cpp",
		".rb":    "ruby",
		".rs":    "rust",
		".sh":    "bash",
		".bash":  "bash",
		".yaml":  "yaml",
		".yml":   "yaml",
		".json":  "json",
		".xml":   "xml",
		".sql":   "sql",
		".proto": "protobuf",
		".tf":    "terraform",
	}
	if l, ok := langs[ext]; ok {
		return l
	}
	return ""
}
