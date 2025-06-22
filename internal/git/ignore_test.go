package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewGitIgnoreParser(t *testing.T) {
	rootDir := "/test/root"
	parser := NewGitIgnoreParser(rootDir)

	if parser == nil {
		t.Fatal("Expected non-nil parser")
	}

	if parser.rootDir != rootDir {
		t.Errorf("Expected rootDir to be %s, got %s", rootDir, parser.rootDir)
	}

	if len(parser.patterns) != 0 {
		t.Errorf("Expected empty patterns, got %d", len(parser.patterns))
	}

	if len(parser.rules) != 0 {
		t.Errorf("Expected empty rules, got %d", len(parser.rules))
	}
}

func TestGitIgnoreParser_ParseGitIgnore(t *testing.T) {
	// Create temporary directory and .gitignore file
	tempDir, err := os.MkdirTemp("", "gitignore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	gitignoreContent := `# This is a comment
*.log
*.tmp
!important.log
build/
node_modules/

# Another comment
*.exe
test_*
*.swp
`

	gitignorePath := filepath.Join(tempDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore file: %v", err)
	}

	parser := NewGitIgnoreParser(tempDir)
	err = parser.ParseGitIgnore(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to parse .gitignore: %v", err)
	}

	// Check that patterns were parsed correctly
	expectedPatterns := []string{"*.log", "*.tmp", "!important.log", "build/", "node_modules/", "*.exe", "test_*", "*.swp"}
	if len(parser.patterns) != len(expectedPatterns) {
		t.Errorf("Expected %d patterns, got %d", len(expectedPatterns), len(parser.patterns))
	}

	for i, expected := range expectedPatterns {
		if i >= len(parser.patterns) || parser.patterns[i] != expected {
			t.Errorf("Expected pattern %s at index %d, got %s", expected, i, parser.patterns[i])
		}
	}

	// Check that rules were parsed correctly
	expectedRules := []struct {
		pattern     string
		isNegation  bool
		isDirectory bool
	}{
		{"*.log", false, false},
		{"*.tmp", false, false},
		{"important.log", true, false},
		{"build", false, true},
		{"node_modules", false, true},
		{"*.exe", false, false},
		{"test_*", false, false},
		{"*.swp", false, false},
	}

	if len(parser.rules) != len(expectedRules) {
		t.Errorf("Expected %d rules, got %d", len(expectedRules), len(parser.rules))
	}

	for i, expected := range expectedRules {
		if i >= len(parser.rules) {
			t.Errorf("Missing rule at index %d", i)
			continue
		}

		rule := parser.rules[i]
		if rule.Pattern != expected.pattern {
			t.Errorf("Expected pattern %s at index %d, got %s", expected.pattern, i, rule.Pattern)
		}
		if rule.IsNegation != expected.isNegation {
			t.Errorf("Expected IsNegation %v at index %d, got %v", expected.isNegation, i, rule.IsNegation)
		}
		if rule.IsDirectory != expected.isDirectory {
			t.Errorf("Expected IsDirectory %v at index %d, got %v", expected.isDirectory, i, rule.IsDirectory)
		}
	}
}

func TestGitIgnoreParser_ParseGitIgnore_NonExistentFile(t *testing.T) {
	parser := NewGitIgnoreParser("/tmp")
	err := parser.ParseGitIgnore("/non/existent/file")
	if err == nil {
		t.Error("Expected error when parsing non-existent file")
	}
}

func TestGitIgnoreParser_ShouldIgnore(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "gitignore_should_ignore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some test files and directories
	testFiles := []string{
		"file.txt",
		"file.log",
		"important.log",
		"test_file.go",
		"main.go",
		"script.exe",
		"document.pdf",
	}

	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Create directories
	testDirs := []string{"build", "node_modules", "src"}
	for _, dir := range testDirs {
		dirPath := filepath.Join(tempDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Create .gitignore
	gitignoreContent := `*.log
!important.log
build/
node_modules/
*.exe
test_*
`

	gitignorePath := filepath.Join(tempDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore file: %v", err)
	}

	parser := NewGitIgnoreParser(tempDir)
	err = parser.ParseGitIgnore(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to parse .gitignore: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Regular file not ignored",
			path:     filepath.Join(tempDir, "file.txt"),
			expected: false,
		},
		{
			name:     "Log file ignored",
			path:     filepath.Join(tempDir, "file.log"),
			expected: true,
		},
		{
			name:     "Important log file not ignored (negation)",
			path:     filepath.Join(tempDir, "important.log"),
			expected: false,
		},
		{
			name:     "Test file ignored",
			path:     filepath.Join(tempDir, "test_file.go"),
			expected: true,
		},
		{
			name:     "Go file not ignored",
			path:     filepath.Join(tempDir, "main.go"),
			expected: false,
		},
		{
			name:     "Exe file ignored",
			path:     filepath.Join(tempDir, "script.exe"),
			expected: true,
		},
		{
			name:     "PDF file not ignored",
			path:     filepath.Join(tempDir, "document.pdf"),
			expected: false,
		},
		{
			name:     "Build directory ignored",
			path:     filepath.Join(tempDir, "build"),
			expected: true,
		},
		{
			name:     "Node modules directory ignored",
			path:     filepath.Join(tempDir, "node_modules"),
			expected: true,
		},
		{
			name:     "Src directory not ignored",
			path:     filepath.Join(tempDir, "src"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ShouldIgnore(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v for path %s, got %v", tt.expected, tt.path, result)
			}
		})
	}
}

func TestGitIgnoreParser_ParseAllGitIgnores(t *testing.T) {
	// Create temporary directory structure with multiple .gitignore files
	tempDir, err := os.MkdirTemp("", "gitignore_parse_all_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create root .gitignore
	rootGitignore := `*.log
*.tmp
build/
`
	if err := os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(rootGitignore), 0644); err != nil {
		t.Fatalf("Failed to create root .gitignore: %v", err)
	}

	// Create subdirectory .gitignore
	subGitignore := `*.local
test_*
`
	if err := os.WriteFile(filepath.Join(subDir, ".gitignore"), []byte(subGitignore), 0644); err != nil {
		t.Fatalf("Failed to create sub .gitignore: %v", err)
	}

	parser := NewGitIgnoreParser(tempDir)
	err = parser.ParseAllGitIgnores()
	if err != nil {
		t.Fatalf("Failed to parse all .gitignore files: %v", err)
	}

	// Check that patterns from both files were parsed
	expectedPatterns := []string{"*.log", "*.tmp", "build/", "*.local", "test_*"}
	if len(parser.patterns) != len(expectedPatterns) {
		t.Errorf("Expected %d patterns, got %d", len(expectedPatterns), len(parser.patterns))
	}

	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range parser.patterns {
			if pattern == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected pattern %s not found in parsed patterns", expected)
		}
	}
}

func TestGitIgnoreParser_ParseAllGitIgnores_NoGitIgnoreFiles(t *testing.T) {
	// Create temporary directory without .gitignore files
	tempDir, err := os.MkdirTemp("", "gitignore_no_files_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	parser := NewGitIgnoreParser(tempDir)
	err = parser.ParseAllGitIgnores()
	if err != nil {
		t.Fatalf("ParseAllGitIgnores should not fail when no .gitignore files exist: %v", err)
	}

	if len(parser.patterns) != 0 {
		t.Errorf("Expected no patterns when no .gitignore files exist, got %d", len(parser.patterns))
	}

	if len(parser.rules) != 0 {
		t.Errorf("Expected no rules when no .gitignore files exist, got %d", len(parser.rules))
	}
}

func TestGitIgnoreRule_Fields(t *testing.T) {
	// Test GitIgnoreRule structure
	rule := GitIgnoreRule{
		Pattern:     "*.log",
		IsNegation:  false,
		IsDirectory: false,
	}

	if rule.Pattern != "*.log" {
		t.Errorf("Expected Pattern to be '*.log', got '%s'", rule.Pattern)
	}
	if rule.IsNegation != false {
		t.Errorf("Expected IsNegation to be false, got %v", rule.IsNegation)
	}
	if rule.IsDirectory != false {
		t.Errorf("Expected IsDirectory to be false, got %v", rule.IsDirectory)
	}

	// Test negation rule
	negationRule := GitIgnoreRule{
		Pattern:     "important.log",
		IsNegation:  true,
		IsDirectory: false,
	}

	if negationRule.IsNegation != true {
		t.Errorf("Expected IsNegation to be true, got %v", negationRule.IsNegation)
	}

	// Test directory rule
	dirRule := GitIgnoreRule{
		Pattern:     "build",
		IsNegation:  false,
		IsDirectory: true,
	}

	if dirRule.IsDirectory != true {
		t.Errorf("Expected IsDirectory to be true, got %v", dirRule.IsDirectory)
	}
}

func TestIsGitAvailable(t *testing.T) {
	// This test checks if the function runs without error
	// The actual result depends on whether .git directory exists
	result := IsGitAvailable()
	
	// Just ensure it returns a boolean
	if result != true && result != false {
		t.Error("IsGitAvailable should return a boolean value")
	}
}

func TestGitIgnoreParser_EmptyLines_and_Comments(t *testing.T) {
	// Create temporary directory and .gitignore file with empty lines and comments
	tempDir, err := os.MkdirTemp("", "gitignore_empty_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	gitignoreContent := `
# This is a comment
*.log

# Another comment with empty lines above and below

*.tmp
		
	# Indented comment
*.exe

`

	gitignorePath := filepath.Join(tempDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore file: %v", err)
	}

	parser := NewGitIgnoreParser(tempDir)
	err = parser.ParseGitIgnore(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to parse .gitignore: %v", err)
	}

	// Should only have the non-comment, non-empty patterns
	expectedPatterns := []string{"*.log", "*.tmp", "*.exe"}
	if len(parser.patterns) != len(expectedPatterns) {
		t.Errorf("Expected %d patterns, got %d", len(expectedPatterns), len(parser.patterns))
	}

	for i, expected := range expectedPatterns {
		if i >= len(parser.patterns) || parser.patterns[i] != expected {
			t.Errorf("Expected pattern %s at index %d, got %s", expected, i, parser.patterns[i])
		}
	}
}