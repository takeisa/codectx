package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewScanner(t *testing.T) {
	scanner := NewScanner("/test/path", true)
	
	if scanner.RootDir != "/test/path" {
		t.Errorf("Expected RootDir to be '/test/path', got '%s'", scanner.RootDir)
	}
	
	if !scanner.IncludeDotfiles {
		t.Error("Expected IncludeDotfiles to be true")
	}
}

func TestScanner_Scan(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "codectx_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test structure
	testFiles := []string{
		"file1.txt",
		"file2.go",
		".hidden",
		"subdir/file3.md",
		"subdir/nested/file4.json",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Test with dotfiles excluded
	scanner := NewScanner(tempDir, false)
	root, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if root == nil {
		t.Fatal("Expected root to be non-nil")
	}

	if !root.IsDir {
		t.Error("Expected root to be a directory")
	}

	// Count files (excluding .hidden)
	paths := scanner.GetRelativePaths(root)
	expectedFiles := 4 // file1.txt, file2.go, subdir/file3.md, subdir/nested/file4.json
	if len(paths) != expectedFiles {
		t.Errorf("Expected %d files, got %d: %v", expectedFiles, len(paths), paths)
	}

	// Test with dotfiles included
	scannerWithDotfiles := NewScanner(tempDir, true)
	rootWithDotfiles, err := scannerWithDotfiles.Scan()
	if err != nil {
		t.Fatalf("Scan with dotfiles failed: %v", err)
	}

	pathsWithDotfiles := scannerWithDotfiles.GetRelativePaths(rootWithDotfiles)
	expectedFilesWithDotfiles := 5 // including .hidden
	if len(pathsWithDotfiles) != expectedFilesWithDotfiles {
		t.Errorf("Expected %d files with dotfiles, got %d: %v", expectedFilesWithDotfiles, len(pathsWithDotfiles), pathsWithDotfiles)
	}
}

func TestScanner_GenerateTree(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "codectx_tree_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test structure
	testStructure := map[string]string{
		"a.txt":           "content",
		"b.go":            "package main",
		"dir1/c.md":       "# Title",
		"dir1/d.json":     "{}",
		"dir2/e.txt":      "text",
		"dir2/sub/f.yaml": "key: value",
	}

	for file, content := range testStructure {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	scanner := NewScanner(tempDir, false)
	root, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	tree := scanner.GenerateTree(root)
	
	// Check if tree contains expected elements
	expectedElements := []string{
		"├── dir1/",
		"│   ├── c.md",
		"│   └── d.json", 
		"├── dir2/",
		"│   ├── sub/",
		"│   │   └── f.yaml",
		"│   └── e.txt",
		"├── a.txt",
		"└── b.go",
	}

	for _, element := range expectedElements {
		if !strings.Contains(tree, element) {
			t.Errorf("Expected tree to contain '%s', but it didn't.\nTree:\n%s", element, tree)
		}
	}
}

func TestScanner_GetRelativePaths(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "codectx_paths_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{
		"root.txt",
		"sub1/file1.go",
		"sub1/sub2/file2.md",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		
		if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	scanner := NewScanner(tempDir, false)
	root, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	paths := scanner.GetRelativePaths(root)
	
	expectedPaths := []string{
		"/root.txt",
		"/sub1/file1.go",
		"/sub1/sub2/file2.md",
	}

	if len(paths) != len(expectedPaths) {
		t.Errorf("Expected %d paths, got %d: %v", len(expectedPaths), len(paths), paths)
	}

	for _, expectedPath := range expectedPaths {
		found := false
		for _, actualPath := range paths {
			if actualPath == expectedPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected path '%s' not found in: %v", expectedPath, paths)
		}
	}
}

func TestScanner_ScanNonExistentDirectory(t *testing.T) {
	scanner := NewScanner("/non/existent/path", false)
	_, err := scanner.Scan()
	
	if err == nil {
		t.Error("Expected error when scanning non-existent directory")
	}
}

func TestScanner_ScanFile(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "codectx_file_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	scanner := NewScanner(tempFile.Name(), false)
	_, err = scanner.Scan()
	
	if err == nil {
		t.Error("Expected error when scanning a file instead of directory")
	}
}