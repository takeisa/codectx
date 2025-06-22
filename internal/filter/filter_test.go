package filter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFilter(t *testing.T) {
	tests := []struct {
		name                string
		extensions          string
		excludePatterns     string
		includeDotfiles     bool
		expectedExtensions  []string
		expectedPatterns    []string
		expectedDotfiles    bool
	}{
		{
			name:               "Empty parameters",
			extensions:         "",
			excludePatterns:    "",
			includeDotfiles:    false,
			expectedExtensions: nil,
			expectedPatterns:   nil,
			expectedDotfiles:   false,
		},
		{
			name:               "Single extension",
			extensions:         "go",
			excludePatterns:    "",
			includeDotfiles:    true,
			expectedExtensions: []string{".go"},
			expectedPatterns:   nil,
			expectedDotfiles:   true,
		},
		{
			name:               "Multiple extensions with dots",
			extensions:         ".js,.ts,py",
			excludePatterns:    "",
			includeDotfiles:    false,
			expectedExtensions: []string{".js", ".ts", ".py"},
			expectedPatterns:   nil,
			expectedDotfiles:   false,
		},
		{
			name:               "Extensions and patterns",
			extensions:         "go,py",
			excludePatterns:    "*.tmp,test_*",
			includeDotfiles:    false,
			expectedExtensions: []string{".go", ".py"},
			expectedPatterns:   []string{"*.tmp", "test_*"},
			expectedDotfiles:   false,
		},
		{
			name:               "Whitespace handling",
			extensions:         " go , md , txt ",
			excludePatterns:    " *.log , temp* ",
			includeDotfiles:    true,
			expectedExtensions: []string{".go", ".md", ".txt"},
			expectedPatterns:   []string{"*.log", "temp*"},
			expectedDotfiles:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFilter(tt.extensions, tt.excludePatterns, tt.includeDotfiles)

			if len(filter.Extensions) != len(tt.expectedExtensions) {
				t.Errorf("Expected %d extensions, got %d", len(tt.expectedExtensions), len(filter.Extensions))
			}

			for i, expected := range tt.expectedExtensions {
				if i >= len(filter.Extensions) || filter.Extensions[i] != expected {
					t.Errorf("Expected extension %s at index %d, got %s", expected, i, filter.Extensions[i])
				}
			}

			if len(filter.ExcludePatterns) != len(tt.expectedPatterns) {
				t.Errorf("Expected %d patterns, got %d", len(tt.expectedPatterns), len(filter.ExcludePatterns))
			}

			for i, expected := range tt.expectedPatterns {
				if i >= len(filter.ExcludePatterns) || filter.ExcludePatterns[i] != expected {
					t.Errorf("Expected pattern %s at index %d, got %s", expected, i, filter.ExcludePatterns[i])
				}
			}

			if filter.IncludeDotfiles != tt.expectedDotfiles {
				t.Errorf("Expected IncludeDotfiles to be %v, got %v", tt.expectedDotfiles, filter.IncludeDotfiles)
			}
		})
	}
}

func TestFilter_ShouldInclude_Dotfiles(t *testing.T) {
	tests := []struct {
		name            string
		includeDotfiles bool
		filePath        string
		expected        bool
	}{
		{
			name:            "Regular file with dotfiles disabled", 
			includeDotfiles: false,
			filePath:        "/path/to/file.txt",
			expected:        true,
		},
		{
			name:            "Dotfile with dotfiles disabled",
			includeDotfiles: false,
			filePath:        "/path/to/.hidden",
			expected:        false,
		},
		{
			name:            "Dotfile with dotfiles enabled",
			includeDotfiles: true,
			filePath:        "/path/to/.hidden",
			expected:        true,
		},
		{
			name:            "Regular file with dotfiles enabled",
			includeDotfiles: true,
			filePath:        "/path/to/file.txt",
			expected:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFilter("", "", tt.includeDotfiles)
			result := filter.ShouldInclude(tt.filePath)
			if result != tt.expected {
				t.Errorf("Expected %v for file %s, got %v", tt.expected, tt.filePath, result)
			}
		})
	}
}

func TestFilter_ShouldInclude_Extensions(t *testing.T) {
	tests := []struct {
		name       string
		extensions string
		filePath   string
		expected   bool
	}{
		{
			name:       "No extension filter - include all",
			extensions: "",
			filePath:   "/path/to/file.anything",
			expected:   true,
		},
		{
			name:       "Matching extension",
			extensions: "go,py",
			filePath:   "/path/to/file.go",
			expected:   true,
		},
		{
			name:       "Non-matching extension",
			extensions: "go,py",
			filePath:   "/path/to/file.txt",
			expected:   false,
		},
		{
			name:       "No extension on file",
			extensions: "go,py",
			filePath:   "/path/to/file",
			expected:   false,
		},
		{
			name:       "Case sensitivity",
			extensions: "go",
			filePath:   "/path/to/file.GO",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFilter(tt.extensions, "", true)
			result := filter.ShouldInclude(tt.filePath)
			if result != tt.expected {
				t.Errorf("Expected %v for file %s with extensions %s, got %v", tt.expected, tt.filePath, tt.extensions, result)
			}
		})
	}
}

func TestFilter_ShouldInclude_ExcludePatterns(t *testing.T) {
	tests := []struct {
		name     string
		patterns string
		filePath string
		expected bool
	}{
		{
			name:     "No exclusion patterns",
			patterns: "",
			filePath: "/path/to/file.txt",
			expected: true,
		},
		{
			name:     "File matches exclusion pattern",
			patterns: "*.tmp,test_*",
			filePath: "/path/to/file.tmp",
			expected: false,
		},
		{
			name:     "File doesn't match exclusion pattern",
			patterns: "*.tmp,test_*",
			filePath: "/path/to/file.txt",
			expected: true,
		},
		{
			name:     "File matches prefix pattern",
			patterns: "test_*",
			filePath: "/path/to/test_file.go",
			expected: false,
		},
		{
			name:     "File matches pattern in path",
			patterns: "temp*",
			filePath: "/project/temp_file.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFilter("", tt.patterns, true)
			result := filter.ShouldInclude(tt.filePath)
			if result != tt.expected {
				t.Errorf("Expected %v for file %s with patterns %s, got %v", tt.expected, tt.filePath, tt.patterns, result)
			}
		})
	}
}

func TestFilter_SetGitTrackedFiles(t *testing.T) {
	filter := NewFilter("", "", true)
	
	// Initially, GitTrackedOnly should be false
	if filter.GitTrackedOnly {
		t.Error("Expected GitTrackedOnly to be false initially")
	}

	trackedFiles := []string{"file1.go", "file2.py", "subdir/file3.md"}
	filter.SetGitTrackedFiles(trackedFiles)

	// After setting tracked files, GitTrackedOnly should be true
	if !filter.GitTrackedOnly {
		t.Error("Expected GitTrackedOnly to be true after setting tracked files")
	}

	if len(filter.GitTrackedFiles) != len(trackedFiles) {
		t.Errorf("Expected %d tracked files, got %d", len(trackedFiles), len(filter.GitTrackedFiles))
	}

	for i, expected := range trackedFiles {
		if filter.GitTrackedFiles[i] != expected {
			t.Errorf("Expected tracked file %s at index %d, got %s", expected, i, filter.GitTrackedFiles[i])
		}
	}
}

func TestFilter_ShouldInclude_GitTrackedFiles(t *testing.T) {
	filter := NewFilter("", "", true)
	trackedFiles := []string{
		"src/main.go",
		"docs/readme.md",
		"config.json",
	}
	filter.SetGitTrackedFiles(trackedFiles)

	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{
			name:     "Tracked file",
			filePath: "src/main.go",
			expected: true,
		},
		{
			name:     "Untracked file",
			filePath: "temp/cache.tmp",
			expected: false,
		},
		{
			name:     "Another tracked file",
			filePath: "config.json",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.ShouldInclude(tt.filePath)
			if result != tt.expected {
				t.Errorf("Expected %v for tracked file %s, got %v", tt.expected, tt.filePath, result)
			}
		})
	}
}

func TestFilter_Integration(t *testing.T) {
	// Create a temporary directory for integration testing
	tempDir, err := os.MkdirTemp("", "filter_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := map[string]string{
		"main.go":        "package main",
		"util.py":        "# Python utility",
		"readme.md":      "# README",
		".hidden":        "hidden content",
		"temp.tmp":       "temporary",
		"test_file.go":   "package test",
		"config.json":    "{}",
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	tests := []struct {
		name            string
		extensions      string
		excludePatterns string
		includeDotfiles bool
		expectedFiles   []string
	}{
		{
			name:            "Go files only",
			extensions:      "go",
			excludePatterns: "",
			includeDotfiles: false,
			expectedFiles:   []string{filepath.Join(tempDir, "main.go"), filepath.Join(tempDir, "test_file.go")},
		},
		{
			name:            "Exclude temp files",
			extensions:      "",
			excludePatterns: "*.tmp,test_*",
			includeDotfiles: false,
			expectedFiles:   []string{filepath.Join(tempDir, "main.go"), filepath.Join(tempDir, "util.py"), filepath.Join(tempDir, "readme.md"), filepath.Join(tempDir, "config.json")},
		},
		{
			name:            "Include dotfiles",
			extensions:      "",
			excludePatterns: "",
			includeDotfiles: true,
			expectedFiles:   []string{filepath.Join(tempDir, "main.go"), filepath.Join(tempDir, "util.py"), filepath.Join(tempDir, "readme.md"), filepath.Join(tempDir, ".hidden"), filepath.Join(tempDir, "temp.tmp"), filepath.Join(tempDir, "test_file.go"), filepath.Join(tempDir, "config.json")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFilter(tt.extensions, tt.excludePatterns, tt.includeDotfiles)
			
			var includedFiles []string
			for filename := range testFiles {
				fullPath := filepath.Join(tempDir, filename)
				if filter.ShouldInclude(fullPath) {
					includedFiles = append(includedFiles, fullPath)
				}
			}

			if len(includedFiles) != len(tt.expectedFiles) {
				t.Errorf("Expected %d files, got %d", len(tt.expectedFiles), len(includedFiles))
				t.Errorf("Expected: %v", tt.expectedFiles)
				t.Errorf("Got: %v", includedFiles)
			}

			// Check if all expected files are included
			for _, expected := range tt.expectedFiles {
				found := false
				for _, included := range includedFiles {
					if included == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected file %s not found in included files", expected)
				}
			}
		})
	}
}