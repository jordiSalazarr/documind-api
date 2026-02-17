package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/ledongthuc/pdf"
)

// PDFParser extracts text content from PDF files.
// It tries pdftotext (poppler) first for reliable extraction,
// falling back to the ledongthuc/pdf Go library.
type PDFParser struct{}

// NewPDFParser creates a new PDF parser.
func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

// SupportedFormats returns the MIME types this parser handles.
func (p *PDFParser) SupportedFormats() []string {
	return []string{"application/pdf"}
}

// Parse extracts text from a PDF file.
func (p *PDFParser) Parse(ctx context.Context, filename string, reader io.Reader) (*ParsedDocument, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF data: %w", err)
	}

	// Try pdftotext first (much more reliable for complex PDFs)
	extracted, pageCount, err := p.extractWithPdftotext(ctx, data)
	if err != nil {
		log.Printf("pdftotext unavailable or failed (%v), falling back to Go library", err)
		extracted, pageCount, err = p.extractWithGoLibrary(ctx, data)
		if err != nil {
			return nil, err
		}
	}

	if strings.TrimSpace(extracted) == "" {
		return nil, fmt.Errorf("no extractable text found in PDF (may be scanned/image-based)")
	}

	title := titleFromFilename(filename)

	return &ParsedDocument{
		Content: extracted,
		Title:   title,
		Metadata: DocumentMetadata{
			SourceFormat: "pdf",
			PageCount:    pageCount,
			WordCount:    countWords(extracted),
		},
	}, nil
}

// extractWithPdftotext uses the poppler pdftotext command for reliable extraction.
func (p *PDFParser) extractWithPdftotext(ctx context.Context, data []byte) (string, int, error) {
	pdftotextPath, err := exec.LookPath("pdftotext")
	if err != nil {
		return "", 0, fmt.Errorf("pdftotext not found: %w", err)
	}

	// Write PDF to temp file (pdftotext needs a file path)
	tmpFile, err := os.CreateTemp("", "documind-pdf-*.pdf")
	if err != nil {
		return "", 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return "", 0, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Extract text with layout preservation
	cmd := exec.CommandContext(ctx, pdftotextPath, "-layout", tmpFile.Name(), "-")
	output, err := cmd.Output()
	if err != nil {
		return "", 0, fmt.Errorf("pdftotext execution failed: %w", err)
	}

	text := strings.TrimSpace(string(output))

	// Get page count from the Go library (pdftotext doesn't expose it easily)
	pageCount := p.getPageCount(data)

	// Format with page markers if we have multiple pages
	if pageCount > 1 {
		text = p.addPageMarkers(text, pageCount)
	}

	return text, pageCount, nil
}

// extractWithGoLibrary uses ledongthuc/pdf as fallback.
func (p *PDFParser) extractWithGoLibrary(ctx context.Context, data []byte) (string, int, error) {
	pdfReader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", 0, fmt.Errorf("failed to open PDF: %w", err)
	}

	pageCount := pdfReader.NumPage()
	var content strings.Builder

	for i := 1; i <= pageCount; i++ {
		if ctx.Err() != nil {
			return "", 0, ctx.Err()
		}

		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}

		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		if content.Len() > 0 {
			content.WriteString("\n\n")
		}
		fmt.Fprintf(&content, "--- Page %d ---\n\n%s", i, text)
	}

	return content.String(), pageCount, nil
}

// getPageCount returns the number of pages using the Go library.
func (p *PDFParser) getPageCount(data []byte) int {
	pdfReader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return 1
	}
	return pdfReader.NumPage()
}

// addPageMarkers splits pdftotext output by form-feed characters (page breaks)
// and adds page markers consistent with the Go library output format.
func (p *PDFParser) addPageMarkers(text string, _ int) string {
	// pdftotext uses form-feed (\f) to separate pages
	pages := strings.Split(text, "\f")

	var result strings.Builder
	for i, page := range pages {
		page = strings.TrimSpace(page)
		if page == "" {
			continue
		}
		if result.Len() > 0 {
			result.WriteString("\n\n")
		}
		fmt.Fprintf(&result, "--- Page %d ---\n\n%s", i+1, page)
	}

	return result.String()
}
