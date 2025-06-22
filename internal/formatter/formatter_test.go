package formatter

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codectx/internal/limits"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name               string
		format             string
		showLineNumbers    bool
		outputPath         string
		expectedFormat     OutputFormat
		expectedError      bool
	}{
		{
			name:            "Text format",
			format:          "text",
			showLineNumbers: true,
			outputPath:      "",
			expectedFormat:  TextFormat,
			expectedError:   false,
		},
		{
			name:            "HTML format",
			format:          "html",
			showLineNumbers: false,
			outputPath:      "",
			expectedFormat:  HTMLFormat,
			expectedError:   false,
		},
		{
			name:            "Markdown format",
			format:          "markdown",
			showLineNumbers: true,
			outputPath:      "",
			expectedFormat:  MarkdownFormat,
			expectedError:   false,
		},
		{
			name:            "JSON format",
			format:          "json",
			showLineNumbers: false,
			outputPath:      "",
			expectedFormat:  JSONFormat,
			expectedError:   false,
		},
		{
			name:            "Case insensitive",
			format:          "TEXT",
			showLineNumbers: true,
			outputPath:      "",
			expectedFormat:  TextFormat,
			expectedError:   false,
		},
		{
			name:            "Invalid format",
			format:          "invalid",
			showLineNumbers: true,
			outputPath:      "",
			expectedFormat:  "",
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sizeLimiter, _ := limits.NewSizeLimiter("1MB", 0)
			formatter, err := NewFormatter(tt.format, tt.showLineNumbers, tt.outputPath, sizeLimiter, nil)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error for format %s, but got none", tt.format)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for format %s: %v", tt.format, err)
				return
			}

			if formatter.Format != tt.expectedFormat {
				t.Errorf("Expected format %s, got %s", tt.expectedFormat, formatter.Format)
			}

			if formatter.ShowLineNumbers != tt.showLineNumbers {
				t.Errorf("Expected ShowLineNumbers %v, got %v", tt.showLineNumbers, formatter.ShowLineNumbers)
			}

			// Clean up
			formatter.Close()
		})
	}
}

func TestNewFormatter_WithOutputFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "formatter_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	sizeLimiter, _ := limits.NewSizeLimiter("1MB", 0)
	formatter, err := NewFormatter("text", true, tempFile.Name(), sizeLimiter, nil)
	if err != nil {
		t.Fatalf("Failed to create formatter with output file: %v", err)
	}
	defer formatter.Close()

	// Verify that the formatter uses the file as writer
	if formatter.Writer == os.Stdout {
		t.Error("Expected formatter to use file writer, but it's using stdout")
	}
}

func TestFormatter_FormatTree_Text(t *testing.T) {
	var buf bytes.Buffer
	sizeLimiter, _ := limits.NewSizeLimiter("1MB", 0)
	formatter := &Formatter{
		Format:          TextFormat,
		ShowLineNumbers: true,
		Writer:          &buf,
		SizeLimiter:     sizeLimiter,
	}

	testTree := `├── dir1/
│   ├── file1.txt
│   └── file2.go
└── dir2/
    └── file3.md`

	err := formatter.FormatTree(testTree)
	if err != nil {
		t.Fatalf("FormatTree failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, testTree) {
		t.Errorf("Expected output to contain tree structure, got: %s", output)
	}
}

func TestFormatter_FormatFileContent_Text(t *testing.T) {
	// Create a temporary file with test content
	tempDir, err := os.MkdirTemp("", "formatter_content_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
	testFile := filepath.Join(tempDir, "test.go")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name            string
		showLineNumbers bool
		expectedContent []string
	}{
		{
			name:            "With line numbers",
			showLineNumbers: true,
			expectedContent: []string{" 1 | package main", " 2 | ", " 3 | import \"fmt\""},
		},
		{
			name:            "Without line numbers",
			showLineNumbers: false,
			expectedContent: []string{"package main", "import \"fmt\"", "func main() {"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			sizeLimiter, _ := limits.NewSizeLimiter("1MB", 0)
			formatter := &Formatter{
				Format:          TextFormat,
				ShowLineNumbers: tt.showLineNumbers,
				Writer:          &buf,
				SizeLimiter:     sizeLimiter,
			}

			relativePath := "/test.go"
			err := formatter.FormatFileContent(testFile, relativePath)
			if err != nil {
				t.Fatalf("FormatFileContent failed: %v", err)
			}

			output := buf.String()

			// Check for file header
			if !strings.Contains(output, relativePath+":") {
				t.Errorf("Expected output to contain file header %s, got: %s", relativePath, output)
			}

			// Check for separator
			if !strings.Contains(output, "---") {
				t.Errorf("Expected output to contain separator, got: %s", output)
			}

			// Check for expected content
			for _, expected := range tt.expectedContent {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, output)
				}
			}
		})
	}
}

func TestFormatter_FormatFileContent_WithSizeLimit(t *testing.T) {
	// Create a temporary file with test content
	tempDir, err := os.MkdirTemp("", "formatter_size_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testContent := strings.Repeat("This is a long line that will exceed the size limit.\n", 100)
	testFile := filepath.Join(tempDir, "large.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a size limiter with a small total size limit
	sizeLimiter, err := limits.NewSizeLimiter("1MB", 500)
	if err != nil {
		t.Fatalf("Failed to create size limiter: %v", err)
	}

	var buf bytes.Buffer
	formatter := &Formatter{
		Format:          TextFormat,
		ShowLineNumbers: false,
		Writer:          &buf,
		SizeLimiter:     sizeLimiter,
	}

	relativePath := "/large.txt"
	err = formatter.FormatFileContent(testFile, relativePath)
	if err != nil {
		t.Fatalf("FormatFileContent failed: %v", err)
	}

	output := buf.String()

	// Should contain truncation message when limit is reached
	if !strings.Contains(output, "truncated") && !strings.Contains(output, "limit") {
		t.Errorf("Expected output to contain truncation message when size limit is reached, got: %s", output)
	}
}

func TestFormatter_FormatFileContent_LargeFile(t *testing.T) {
	// Create a temporary file that exceeds file size limit
	tempDir, err := os.MkdirTemp("", "formatter_large_file_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a large file (simulate by creating small size limiter)
	testContent := strings.Repeat("Large file content\n", 1000)
	testFile := filepath.Join(tempDir, "large.txt")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a size limiter with very small file size limit
	sizeLimiter, err := limits.NewSizeLimiter("1KB", 0)
	if err != nil {
		t.Fatalf("Failed to create size limiter: %v", err)
	}

	var buf bytes.Buffer
	formatter := &Formatter{
		Format:          TextFormat,
		ShowLineNumbers: false,
		Writer:          &buf,
		SizeLimiter:     sizeLimiter,
	}

	relativePath := "/large.txt"
	err = formatter.FormatFileContent(testFile, relativePath)
	if err != nil {
		t.Fatalf("FormatFileContent failed: %v", err)
	}

	output := buf.String()

	// Should contain file too large message
	if !strings.Contains(output, "too large") && !strings.Contains(output, "exceeds") {
		t.Errorf("Expected output to contain 'too large' message for large file, got: %s", output)
	}
}

func TestFormatter_Close(t *testing.T) {
	// Test with stdout (should not fail)
	var buf bytes.Buffer
	sizeLimiter, _ := limits.NewSizeLimiter("1MB", 0)
	formatter1 := &Formatter{
		Format:          TextFormat,
		ShowLineNumbers: true,
		Writer:          &buf,
		SizeLimiter:     sizeLimiter,
	}

	err := formatter1.Close()
	if err != nil {
		t.Errorf("Close with buffer writer should not fail, got: %v", err)
	}

	// Test with file writer
	tempFile, err := os.CreateTemp("", "formatter_close_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	formatter2, err := NewFormatter("text", true, tempFile.Name(), sizeLimiter, nil)
	if err != nil {
		t.Fatalf("Failed to create formatter: %v", err)
	}

	err = formatter2.Close()
	if err != nil {
		t.Errorf("Close with file writer should not fail, got: %v", err)
	}
}

func TestOutputFormatConstants(t *testing.T) {
	// Test that format constants are correctly defined
	if TextFormat != "text" {
		t.Errorf("Expected TextFormat to be 'text', got '%s'", TextFormat)
	}
	if HTMLFormat != "html" {
		t.Errorf("Expected HTMLFormat to be 'html', got '%s'", HTMLFormat)
	}
	if MarkdownFormat != "markdown" {
		t.Errorf("Expected MarkdownFormat to be 'markdown', got '%s'", MarkdownFormat)
	}
	if JSONFormat != "json" {
		t.Errorf("Expected JSONFormat to be 'json', got '%s'", JSONFormat)
	}
}