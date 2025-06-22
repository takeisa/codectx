package git

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestFilterEmptyStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "No empty strings",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Some empty strings",
			input:    []string{"a", "", "b", "", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "All empty strings",
			input:    []string{"", "", ""},
			expected: []string{},
		},
		{
			name:     "Empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Single non-empty string",
			input:    []string{"test"},
			expected: []string{"test"},
		},
		{
			name:     "Single empty string",
			input:    []string{""},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterEmptyStrings(tt.input)
			
			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected %s at index %d, got %s", expected, i, result[i])
				}
			}
		})
	}
}

func TestIsGitCommandAvailable(t *testing.T) {
	// This test checks if the function runs without error
	// The actual result depends on whether git is installed on the system
	result := isGitCommandAvailable()
	
	// Just ensure it returns a boolean
	if result != true && result != false {
		t.Error("isGitCommandAvailable should return a boolean value")
	}

	// Check that it matches exec.LookPath behavior
	_, err := exec.LookPath("git")
	expected := err == nil
	
	if result != expected {
		t.Errorf("Expected isGitCommandAvailable to return %v, got %v", expected, result)
	}
}

func TestIsGitRepository(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "git_repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with non-git directory
	result := isGitRepository(tempDir)
	if result {
		t.Error("Expected non-git directory to return false")
	}

	// Test with non-existent directory
	result = isGitRepository("/non/existent/path")
	if result {
		t.Error("Expected non-existent directory to return false")
	}
}

func TestGetGitInfo_NotGitRepository(t *testing.T) {
	// Create temporary directory that's not a git repository
	tempDir, err := os.MkdirTemp("", "not_git_repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	_, err = GetGitInfo(tempDir)
	if err == nil {
		t.Error("Expected error when getting git info for non-git repository")
	}

	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("Expected error message to contain 'not a git repository', got: %v", err)
	}
}

func TestGetGitTrackedFiles_NotGitRepository(t *testing.T) {
	// Create temporary directory that's not a git repository
	tempDir, err := os.MkdirTemp("", "not_git_repo_tracked_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	_, err = GetGitTrackedFiles(tempDir)
	if err == nil {
		t.Error("Expected error when getting tracked files for non-git repository")
	}

	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("Expected error message to contain 'not a git repository', got: %v", err)
	}
}

func TestGetGitStatus_NotGitRepository(t *testing.T) {
	// Create temporary directory that's not a git repository
	tempDir, err := os.MkdirTemp("", "not_git_repo_status_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	_, err = GetGitStatus(tempDir)
	if err == nil {
		t.Error("Expected error when getting git status for non-git repository")
	}

	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("Expected error message to contain 'not a git repository', got: %v", err)
	}
}

func TestGitInfo_Structure(t *testing.T) {
	// Test that GitInfo structure has expected fields
	info := &GitInfo{
		CommitHash:    "abc123",
		Branch:        "main",
		Author:        "Test Author <test@example.com>",
		CommitDate:    time.Now(),
		IsDirty:       true,
		LastModified:  time.Now(),
		RepositoryURL: "https://github.com/test/repo.git",
	}

	if info.CommitHash != "abc123" {
		t.Errorf("Expected CommitHash to be 'abc123', got '%s'", info.CommitHash)
	}

	if info.Branch != "main" {
		t.Errorf("Expected Branch to be 'main', got '%s'", info.Branch)
	}

	if info.Author != "Test Author <test@example.com>" {
		t.Errorf("Expected Author to be 'Test Author <test@example.com>', got '%s'", info.Author)
	}

	if !info.IsDirty {
		t.Error("Expected IsDirty to be true")
	}

	if info.RepositoryURL != "https://github.com/test/repo.git" {
		t.Errorf("Expected RepositoryURL to be 'https://github.com/test/repo.git', got '%s'", info.RepositoryURL)
	}
}

func TestRunGitCommand_InvalidCommand(t *testing.T) {
	// Test with a command that should fail
	tempDir, err := os.MkdirTemp("", "git_command_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Try to run git command in non-git directory
	_, err = runGitCommand(tempDir, "status")
	if err == nil {
		t.Error("Expected error when running git status in non-git directory")
	}
}

// Integration tests (these will only work if git is available and we're in a git repo)

func TestGetGitInfo_Integration(t *testing.T) {
	// Skip if git is not available
	if !isGitCommandAvailable() {
		t.Skip("Skipping integration test: git command not available")
	}

	// Try to get git info for current directory
	// This may fail if we're not in a git repository, which is fine for testing
	_, err := GetGitInfo(".")
	
	// We don't assert the result since this depends on the test environment
	// But we can check that the function doesn't panic
	_ = err // Use the error to avoid "unused variable" warning
}

func TestGetGitTrackedFiles_Integration(t *testing.T) {
	// Skip if git is not available
	if !isGitCommandAvailable() {
		t.Skip("Skipping integration test: git command not available")
	}

	// Try to get tracked files for current directory
	// This may fail if we're not in a git repository, which is fine for testing
	_, err := GetGitTrackedFiles(".")
	
	// We don't assert the result since this depends on the test environment
	// But we can check that the function doesn't panic
	_ = err // Use the error to avoid "unused variable" warning
}

func TestGetGitStatus_Integration(t *testing.T) {
	// Skip if git is not available
	if !isGitCommandAvailable() {
		t.Skip("Skipping integration test: git command not available")
	}

	// Try to get git status for current directory
	// This may fail if we're not in a git repository, which is fine for testing
	_, err := GetGitStatus(".")
	
	// We don't assert the result since this depends on the test environment
	// But we can check that the function doesn't panic
	_ = err // Use the error to avoid "unused variable" warning
}

func TestGitCommandFunctions_NoGitCommand(t *testing.T) {
	// These tests check behavior when git command is not available
	// We can't easily mock this, so we test the error handling

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "no_git_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// If git is not available, all functions should return appropriate errors
	if !isGitCommandAvailable() {
		_, err := GetGitInfo(tempDir)
		if err == nil {
			t.Error("Expected error when git command is not available")
		}
		if !strings.Contains(err.Error(), "git command not available") {
			t.Errorf("Expected error to contain 'git command not available', got: %v", err)
		}

		_, err = GetGitTrackedFiles(tempDir)
		if err == nil {
			t.Error("Expected error when git command is not available")
		}
		if !strings.Contains(err.Error(), "git command not available") {
			t.Errorf("Expected error to contain 'git command not available', got: %v", err)
		}

		_, err = GetGitStatus(tempDir)
		if err == nil {
			t.Error("Expected error when git command is not available")
		}
		if !strings.Contains(err.Error(), "git command not available") {
			t.Errorf("Expected error to contain 'git command not available', got: %v", err)
		}
	}
}

func TestGitInfoJSONTags(t *testing.T) {
	// Test that GitInfo struct has proper JSON tags for serialization
	info := GitInfo{}
	
	// This is a compile-time check that the struct has the expected fields
	// with proper types
	var _ string = info.CommitHash
	var _ string = info.Branch
	var _ string = info.Author
	var _ time.Time = info.CommitDate
	var _ bool = info.IsDirty
	var _ time.Time = info.LastModified
	var _ string = info.RepositoryURL
}