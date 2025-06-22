package stats

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codectx/internal/utils"
)

func TestNewStatsCollector(t *testing.T) {
	collector := NewStatsCollector()

	if collector == nil {
		t.Fatal("Expected non-nil stats collector")
	}

	if collector.TotalFiles != 0 {
		t.Errorf("Expected TotalFiles to be 0, got %d", collector.TotalFiles)
	}

	if collector.TotalDirectories != 0 {
		t.Errorf("Expected TotalDirectories to be 0, got %d", collector.TotalDirectories)
	}

	if collector.TotalSize != 0 {
		t.Errorf("Expected TotalSize to be 0, got %d", collector.TotalSize)
	}

	if collector.TextFiles != 0 {
		t.Errorf("Expected TextFiles to be 0, got %d", collector.TextFiles)
	}

	if collector.BinaryFiles != 0 {
		t.Errorf("Expected BinaryFiles to be 0, got %d", collector.BinaryFiles)
	}

	if collector.EstimatedTokens != 0 {
		t.Errorf("Expected EstimatedTokens to be 0, got %d", collector.EstimatedTokens)
	}

	// Check that start time is recent
	if time.Since(collector.StartTime) > time.Second {
		t.Error("Expected StartTime to be recent")
	}
}

func TestStatsCollector_AddDirectory(t *testing.T) {
	collector := NewStatsCollector()
	
	collector.AddDirectory("/path/to/dir1")
	collector.AddDirectory("/path/to/dir2")
	collector.AddDirectory("/path/to/dir3")

	if collector.TotalDirectories != 3 {
		t.Errorf("Expected TotalDirectories to be 3, got %d", collector.TotalDirectories)
	}
}

func TestStatsCollector_AddFile(t *testing.T) {
	// Create temporary files for testing
	tempDir, err := os.MkdirTemp("", "stats_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a text file
	textFile := filepath.Join(tempDir, "test.txt")
	textContent := "This is a test file with some content."
	if err := os.WriteFile(textFile, []byte(textContent), 0644); err != nil {
		t.Fatalf("Failed to create text file: %v", err)
	}

	// Create a binary file (simulate by creating a file with null bytes)
	binaryFile := filepath.Join(tempDir, "test.bin")
	binaryContent := []byte{0x00, 0x01, 0x02, 0x03, 0xFF}
	if err := os.WriteFile(binaryFile, binaryContent, 0644); err != nil {
		t.Fatalf("Failed to create binary file: %v", err)
	}

	collector := NewStatsCollector()

	// Add text file
	err = collector.AddFile(textFile, true)
	if err != nil {
		t.Fatalf("Failed to add text file: %v", err)
	}

	// Add binary file
	err = collector.AddFile(binaryFile, false)
	if err != nil {
		t.Fatalf("Failed to add binary file: %v", err)
	}

	// Check stats
	if collector.TotalFiles != 2 {
		t.Errorf("Expected TotalFiles to be 2, got %d", collector.TotalFiles)
	}

	if collector.TextFiles != 1 {
		t.Errorf("Expected TextFiles to be 1, got %d", collector.TextFiles)
	}

	if collector.BinaryFiles != 1 {
		t.Errorf("Expected BinaryFiles to be 1, got %d", collector.BinaryFiles)
	}

	expectedSize := int64(len(textContent) + len(binaryContent))
	if collector.TotalSize != expectedSize {
		t.Errorf("Expected TotalSize to be %d, got %d", expectedSize, collector.TotalSize)
	}

	// Text file should contribute to token count
	if collector.EstimatedTokens == 0 {
		t.Error("Expected EstimatedTokens to be greater than 0 for text file")
	}
}

func TestStatsCollector_GetProcessingTime(t *testing.T) {
	collector := NewStatsCollector()
	
	// Sleep for a short time to ensure processing time is measurable
	time.Sleep(10 * time.Millisecond)
	
	processingTime := collector.GetProcessingTime()
	if processingTime <= 0 {
		t.Errorf("Expected processing time to be positive, got %f", processingTime)
	}

	if processingTime > 1.0 {
		t.Errorf("Expected processing time to be less than 1 second, got %f", processingTime)
	}
}

func TestIsTextFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "text_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test cases
	tests := []struct {
		name        string
		content     []byte
		expected    bool
		expectError bool
	}{
		{
			name:        "Plain text",
			content:     []byte("This is plain text content."),
			expected:    true,
			expectError: false,
		},
		{
			name:        "Code file",
			content:     []byte("package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}"),
			expected:    true,
			expectError: false,
		},
		{
			name:        "Binary with null bytes",
			content:     []byte{0x00, 0x01, 0x02, 0x03, 0xFF},
			expected:    false,
			expectError: false,
		},
		{
			name:        "UTF-8 text",
			content:     []byte("Hello 世界! 🌍"),
			expected:    true,
			expectError: false,
		},
		{
			name:        "Empty file",
			content:     []byte(""),
			expected:    true,
			expectError: false,
		},
		{
			name:        "Mixed content with null",
			content:     []byte("Some text\x00more text"),
			expected:    false,
			expectError: false,
		},
		{
			name:        "Markdown file",
			content:     []byte("# Title\n\nThis is **markdown** content."),
			expected:    true,
			expectError: false,
		},
		{
			name:        "Go source file",
			content:     []byte("package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}"),
			expected:    true,
			expectError: false,
		},
		{
			name:        "JSON file",
			content:     []byte("{\"name\": \"test\", \"value\": 42}"),
			expected:    true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := "test_" + strings.ReplaceAll(tt.name, " ", "_")
			// Add appropriate extension for specific test cases
			switch tt.name {
			case "Markdown file":
				filename += ".md"
			case "Go source file":
				filename += ".go"
			case "JSON file":
				filename += ".json"
			case "Binary with null bytes", "Mixed content with null":
				// No extension for these cases to test content-based detection
				break
			default:
				filename += ".txt"
			}
			
			testFile := filepath.Join(tempDir, filename)
			if err := os.WriteFile(testFile, tt.content, 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result, err := utils.IsTextFile(testFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.name)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.name, err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %v for %s, got %v", tt.expected, tt.name, result)
			}
		})
	}
}

func TestIsTextFileByExtension(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "extension_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test cases for extension-based detection
	tests := []struct {
		filename string
		content  []byte
		expected bool
	}{
		{"test.md", []byte("# Markdown content"), true},
		{"README.md", []byte("# README file"), true},
		{"test.txt", []byte("Plain text"), true},
		{"config.json", []byte("{\"key\": \"value\"}"), true},
		{"script.py", []byte("print('Hello World')"), true},
		{"main.go", []byte("package main"), true},
		{"style.css", []byte("body { color: red; }"), true},
		{"index.html", []byte("<html><body>Hello</body></html>"), true},
		{"app.js", []byte("console.log('Hello');"), true},
		{"data.xml", []byte("<?xml version=\"1.0\"?><root></root>"), true},
		{"config.yaml", []byte("key: value"), true},
		{"script.sh", []byte("#!/bin/bash\necho hello"), true},
		{"Dockerfile", []byte("FROM ubuntu"), true},
		{".gitignore", []byte("*.log\n*.tmp"), true},
		{"binary.bin", []byte{0x00, 0x01, 0x02, 0x03}, false}, // No known text extension
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			testFile := filepath.Join(tempDir, tt.filename)
			if err := os.WriteFile(testFile, tt.content, 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result, err := utils.IsTextFile(testFile)
			if err != nil {
				t.Fatalf("IsTextFile failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v for %s, got %v", tt.expected, tt.filename, result)
			}
		})
	}
}

func TestEstimateTokens(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "token_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		filename    string
		content     string
		minTokens   int
		maxTokens   int
	}{
		{
			name:      "Go code",
			filename:  "test.go",
			content:   "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}",
			minTokens: 5,
			maxTokens: 20,
		},
		{
			name:      "Python code",
			filename:  "test.py",
			content:   "def hello():\n    print(\"Hello, World!\")\n\nhello()",
			minTokens: 4,
			maxTokens: 15,
		},
		{
			name:      "JSON data",
			filename:  "test.json",
			content:   `{"name": "test", "value": 42, "active": true}`,
			minTokens: 3,
			maxTokens: 12,
		},
		{
			name:      "Markdown text",
			filename:  "test.md",
			content:   "# Title\n\nThis is a **markdown** document with some content.",
			minTokens: 6,
			maxTokens: 15,
		},
		{
			name:      "Plain text",
			filename:  "test.txt",
			content:   "This is plain text with multiple words and sentences.",
			minTokens: 8,
			maxTokens: 15,
		},
		{
			name:      "Empty file",
			filename:  "empty.txt",
			content:   "",
			minTokens: 0,
			maxTokens: 0,
		},
		{
			name:      "Single line",
			filename:  "single.txt",
			content:   "Single line",
			minTokens: 1,
			maxTokens: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tempDir, tt.filename)
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			tokens, err := EstimateTokens(testFile)
			if err != nil {
				t.Fatalf("EstimateTokens failed: %v", err)
			}

			if tokens < tt.minTokens || tokens > tt.maxTokens {
				t.Errorf("Expected tokens between %d and %d for %s, got %d", tt.minTokens, tt.maxTokens, tt.name, tokens)
			}
		})
	}
}

func TestEstimateCodeLineTokens(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		minTokens int
		maxTokens int
	}{
		{
			name:      "Simple assignment",
			line:      "x := 42",
			minTokens: 2,
			maxTokens: 5,
		},
		{
			name:      "Function call",
			line:      "fmt.Println(\"Hello, World!\")",
			minTokens: 3,
			maxTokens: 8,
		},
		{
			name:      "Complex expression",
			line:      "result := (a + b) * c / d",
			minTokens: 10,
			maxTokens: 20,
		},
		{
			name:      "Comment line",
			line:      "// This is a comment",
			minTokens: 0,
			maxTokens: 2,
		},
		{
			name:      "Empty line",
			line:      "",
			minTokens: 0,
			maxTokens: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := estimateCodeLineTokens(tt.line)
			if tokens < tt.minTokens || tokens > tt.maxTokens {
				t.Errorf("Expected tokens between %d and %d for line '%s', got %d", tt.minTokens, tt.maxTokens, tt.line, tokens)
			}
		})
	}
}

func TestEstimateDataLineTokens(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		minTokens int
		maxTokens int
	}{
		{
			name:      "JSON object",
			line:      `{"key": "value", "number": 42}`,
			minTokens: 3,
			maxTokens: 6,
		},
		{
			name:      "JSON array",
			line:      `[1, 2, 3, 4, 5]`,
			minTokens: 4,
			maxTokens: 8,
		},
		{
			name:      "Simple key-value",
			line:      `"name": "test"`,
			minTokens: 2,
			maxTokens: 4,
		},
		{
			name:      "Empty object",
			line:      `{}`,
			minTokens: 0,
			maxTokens: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := estimateDataLineTokens(tt.line)
			if tokens < tt.minTokens || tokens > tt.maxTokens {
				t.Errorf("Expected tokens between %d and %d for line '%s', got %d", tt.minTokens, tt.maxTokens, tt.line, tokens)
			}
		})
	}
}

func TestEstimateTextLineTokens(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		minTokens int
		maxTokens int
	}{
		{
			name:      "Simple sentence",
			line:      "This is a simple sentence.",
			minTokens: 4,
			maxTokens: 8,
		},
		{
			name:      "Complex sentence",
			line:      "The quick brown fox jumps over the lazy dog!",
			minTokens: 8,
			maxTokens: 15,
		},
		{
			name:      "Single word",
			line:      "Word",
			minTokens: 1,
			maxTokens: 2,
		},
		{
			name:      "Empty line",
			line:      "",
			minTokens: 0,
			maxTokens: 1,
		},
		{
			name:      "Punctuation heavy",
			line:      "Hello, world! How are you? I'm fine, thanks.",
			minTokens: 6,
			maxTokens: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := estimateTextLineTokens(tt.line)
			if tokens < tt.minTokens || tokens > tt.maxTokens {
				t.Errorf("Expected tokens between %d and %d for line '%s', got %d", tt.minTokens, tt.maxTokens, tt.line, tokens)
			}
		})
	}
}

func TestCollectStats(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "collect_stats_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test structure
	testFiles := map[string][]byte{
		"file1.txt":      []byte("This is a text file."),
		"file2.go":       []byte("package main\n\nfunc main() {}"),
		"subdir/file3.md": []byte("# Markdown\n\nContent here."),
		"binary.bin":     {0x00, 0x01, 0x02, 0x03},
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)
		dir := filepath.Dir(fullPath)
		
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	stats, err := CollectStats(tempDir)
	if err != nil {
		t.Fatalf("CollectStats failed: %v", err)
	}

	// Check basic stats
	if stats.TotalFiles != 4 {
		t.Errorf("Expected 4 files, got %d", stats.TotalFiles)
	}

	if stats.TotalDirectories != 2 { // tempDir + subdir
		t.Errorf("Expected 2 directories, got %d", stats.TotalDirectories)
	}

	if stats.TextFiles != 3 {
		t.Errorf("Expected 3 text files, got %d", stats.TextFiles)
	}

	if stats.BinaryFiles != 1 {
		t.Errorf("Expected 1 binary file, got %d", stats.BinaryFiles)
	}

	if stats.TotalSize == 0 {
		t.Error("Expected TotalSize to be greater than 0")
	}

	if stats.EstimatedTokens == 0 {
		t.Error("Expected EstimatedTokens to be greater than 0")
	}

	if stats.GetProcessingTime() <= 0 {
		t.Error("Expected processing time to be positive")
	}
}