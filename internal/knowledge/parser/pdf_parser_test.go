package parser

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestPDFParser_SimplePDF(t *testing.T) {
	expectedText := "Hello World! This is a test PDF document with multiple sentences. " +
		"It should be properly extracted by the parser. " +
		"The quick brown fox jumps over the lazy dog."

	pdfData := createMinimalPDF(expectedText)

	parser := NewPDFParser()
	doc, err := parser.Parse(context.Background(), "test_document.pdf", bytes.NewReader(pdfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	t.Logf("Extracted content (%d chars): %q", len(doc.Content), truncate(doc.Content, 300))

	if !strings.Contains(doc.Content, "Hello World") {
		t.Errorf("Expected content to contain 'Hello World', got: %s", truncate(doc.Content, 200))
	}
	if doc.Title != "test document" {
		t.Errorf("Expected title 'test document', got: %q", doc.Title)
	}
	if doc.Metadata.SourceFormat != "pdf" {
		t.Errorf("Expected source format 'pdf', got: %q", doc.Metadata.SourceFormat)
	}
	if doc.Metadata.PageCount < 1 {
		t.Errorf("Expected at least 1 page, got: %d", doc.Metadata.PageCount)
	}
}

func TestPDFParser_PdftotextFallback(t *testing.T) {
	expectedText := "This text should be extracted correctly."
	pdfData := createMinimalPDF(expectedText)

	parser := NewPDFParser()

	// Test pdftotext path directly
	t.Run("pdftotext_extraction", func(t *testing.T) {
		text, pageCount, err := parser.extractWithPdftotext(context.Background(), pdfData)
		if err != nil {
			t.Skipf("pdftotext not available: %v", err)
		}
		t.Logf("pdftotext: %d pages, %d chars: %q", pageCount, len(text), truncate(text, 200))
		if !strings.Contains(text, "extracted correctly") {
			t.Errorf("pdftotext failed to extract expected text")
		}
	})

	// Test Go library path directly
	t.Run("go_library_extraction", func(t *testing.T) {
		text, pageCount, err := parser.extractWithGoLibrary(context.Background(), pdfData)
		if err != nil {
			t.Fatalf("Go library failed: %v", err)
		}
		t.Logf("Go library: %d pages, %d chars: %q", pageCount, len(text), truncate(text, 200))
		if !strings.Contains(text, "extracted correctly") {
			t.Errorf("Go library failed to extract expected text")
		}
	})
}

func TestPDFParser_EmptyPDF(t *testing.T) {
	// Create a PDF with no text content
	pdfData := createMinimalPDFEmpty()

	parser := NewPDFParser()
	_, err := parser.Parse(context.Background(), "empty.pdf", bytes.NewReader(pdfData))
	if err == nil {
		t.Error("Expected error for PDF with no extractable text")
	}
	if err != nil && !strings.Contains(err.Error(), "no extractable text") {
		t.Logf("Got expected error: %v", err)
	}
}

func TestPDFParser_PdftotextAvailability(t *testing.T) {
	_, err := exec.LookPath("pdftotext")
	if err != nil {
		t.Log("pdftotext NOT available — parser will use Go library fallback")
	} else {
		t.Log("pdftotext IS available — parser will use it as primary extractor")
	}
}

// --- helpers ---

func createMinimalPDF(text string) []byte {
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")

	obj1Offset := buf.Len()
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	obj2Offset := buf.Len()
	buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	obj4Offset := buf.Len()
	buf.WriteString("4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	streamContent := fmt.Sprintf("BT /F1 12 Tf 72 720 Td (%s) Tj ET", escapePDFString(text))
	obj5Offset := buf.Len()
	fmt.Fprintf(&buf, "5 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(streamContent), streamContent)

	obj3Offset := buf.Len()
	buf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 5 0 R /Resources << /Font << /F1 4 0 R >> >> >>\nendobj\n")

	xrefOffset := buf.Len()
	buf.WriteString("xref\n0 6\n")
	fmt.Fprintf(&buf, "0000000000 65535 f \n")
	fmt.Fprintf(&buf, "%010d 00000 n \n", obj1Offset)
	fmt.Fprintf(&buf, "%010d 00000 n \n", obj2Offset)
	fmt.Fprintf(&buf, "%010d 00000 n \n", obj3Offset)
	fmt.Fprintf(&buf, "%010d 00000 n \n", obj4Offset)
	fmt.Fprintf(&buf, "%010d 00000 n \n", obj5Offset)
	fmt.Fprintf(&buf, "trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefOffset)

	return buf.Bytes()
}

func createMinimalPDFEmpty() []byte {
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")

	obj1Offset := buf.Len()
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	obj2Offset := buf.Len()
	buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	// Empty content stream
	obj4Offset := buf.Len()
	buf.WriteString("4 0 obj\n<< /Length 0 >>\nstream\n\nendstream\nendobj\n")

	obj3Offset := buf.Len()
	buf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << >> >>\nendobj\n")

	xrefOffset := buf.Len()
	buf.WriteString("xref\n0 5\n")
	fmt.Fprintf(&buf, "0000000000 65535 f \n")
	fmt.Fprintf(&buf, "%010d 00000 n \n", obj1Offset)
	fmt.Fprintf(&buf, "%010d 00000 n \n", obj2Offset)
	fmt.Fprintf(&buf, "%010d 00000 n \n", obj3Offset)
	fmt.Fprintf(&buf, "%010d 00000 n \n", obj4Offset)
	fmt.Fprintf(&buf, "trailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefOffset)

	return buf.Bytes()
}

func escapePDFString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// TestPDFParser_RealPDFFile tests with an actual PDF file if one exists at a known path.
func TestPDFParser_RealPDFFile(t *testing.T) {
	// Look for any PDF in common locations
	paths := []string{
		os.Getenv("TEST_PDF_PATH"), // Allow custom path via env var
	}

	var testFile string
	for _, p := range paths {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			testFile = p
			break
		}
	}

	if testFile == "" {
		t.Skip("No test PDF file found. Set TEST_PDF_PATH env var to test with a real PDF.")
	}

	f, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open %s: %v", testFile, err)
	}
	defer f.Close()

	parser := NewPDFParser()
	doc, err := parser.Parse(context.Background(), testFile, f)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	t.Logf("File: %s", testFile)
	t.Logf("Title: %s", doc.Title)
	t.Logf("Pages: %d", doc.Metadata.PageCount)
	t.Logf("Words: %d", doc.Metadata.WordCount)
	t.Logf("Content preview: %s", truncate(doc.Content, 500))

	if doc.Metadata.WordCount < 10 {
		t.Errorf("Suspiciously low word count (%d) — extraction may have failed", doc.Metadata.WordCount)
	}
}
