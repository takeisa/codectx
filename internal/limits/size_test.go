package limits

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int64
		expectError bool
	}{
		{
			name:        "Empty string",
			input:       "",
			expected:    0,
			expectError: false,
		},
		{
			name:        "Bytes only",
			input:       "1024",
			expected:    1024,
			expectError: false,
		},
		{
			name:        "Bytes with B unit",
			input:       "512B",
			expected:    512,
			expectError: false,
		},
		{
			name:        "Kilobytes",
			input:       "1KB",
			expected:    1024,
			expectError: false,
		},
		{
			name:        "Megabytes",
			input:       "1MB",
			expected:    1024 * 1024,
			expectError: false,
		},
		{
			name:        "Gigabytes",
			input:       "1GB",
			expected:    1024 * 1024 * 1024,
			expectError: false,
		},
		{
			name:        "Multiple KB",
			input:       "500KB",
			expected:    500 * 1024,
			expectError: false,
		},
		{
			name:        "Multiple MB",
			input:       "10MB",
			expected:    10 * 1024 * 1024,
			expectError: false,
		},
		{
			name:        "Lowercase unit",
			input:       "1mb",
			expected:    1024 * 1024,
			expectError: false,
		},
		{
			name:        "Mixed case unit",
			input:       "1Mb",
			expected:    1024 * 1024,
			expectError: false,
		},
		{
			name:        "With whitespace",
			input:       " 5 MB ",
			expected:    5 * 1024 * 1024,
			expectError: false,
		},
		{
			name:        "Zero value",
			input:       "0MB",
			expected:    0,
			expectError: false,
		},
		{
			name:        "Invalid unit",
			input:       "1TB",
			expected:    0,
			expectError: true,
		},
		{
			name:        "Invalid number",
			input:       "abcMB",
			expected:    0,
			expectError: true,
		},
		{
			name:        "Negative number",
			input:       "-1MB",
			expected:    0,
			expectError: true,
		},
		{
			name:        "Invalid format",
			input:       "1.5MB",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSize(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input %s, but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", tt.input, err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %d for input %s, got %d", tt.expected, tt.input, result)
			}
		})
	}
}

func TestNewSizeLimiter(t *testing.T) {
	tests := []struct {
		name              string
		maxFileSize       string
		maxTotalSize      int64
		expectedFileSize  int64
		expectedTotalSize int64
		expectError       bool
	}{
		{
			name:              "Default values",
			maxFileSize:       "",
			maxTotalSize:      0,
			expectedFileSize:  1024 * 1024, // Default 1MB
			expectedTotalSize: 0,
			expectError:       false,
		},
		{
			name:              "Custom file size",
			maxFileSize:       "500KB",
			maxTotalSize:      10000,
			expectedFileSize:  500 * 1024,
			expectedTotalSize: 10000,
			expectError:       false,
		},
		{
			name:              "Large limits",
			maxFileSize:       "10MB",
			maxTotalSize:      1000000,
			expectedFileSize:  10 * 1024 * 1024,
			expectedTotalSize: 1000000,
			expectError:       false,
		},
		{
			name:              "Invalid file size",
			maxFileSize:       "invalid",
			maxTotalSize:      0,
			expectedFileSize:  0,
			expectedTotalSize: 0,
			expectError:       true,
		},
		{
			name:              "Zero total size",
			maxFileSize:       "1MB",
			maxTotalSize:      0,
			expectedFileSize:  1024 * 1024,
			expectedTotalSize: 0,
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter, err := NewSizeLimiter(tt.maxFileSize, tt.maxTotalSize)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for maxFileSize %s, but got none", tt.maxFileSize)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for maxFileSize %s: %v", tt.maxFileSize, err)
				return
			}

			if limiter.MaxFileSize != tt.expectedFileSize {
				t.Errorf("Expected MaxFileSize %d, got %d", tt.expectedFileSize, limiter.MaxFileSize)
			}

			if limiter.MaxTotalSize != tt.expectedTotalSize {
				t.Errorf("Expected MaxTotalSize %d, got %d", tt.expectedTotalSize, limiter.MaxTotalSize)
			}

			if limiter.CurrentTotalSize != 0 {
				t.Errorf("Expected CurrentTotalSize to be 0, got %d", limiter.CurrentTotalSize)
			}
		})
	}
}

func TestSizeLimiter_CheckFileSize(t *testing.T) {
	// Create temporary files for testing
	tempDir, err := os.MkdirTemp("", "size_limiter_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a small file
	smallFile := filepath.Join(tempDir, "small.txt")
	smallContent := []byte("This is a small file.")
	if err := os.WriteFile(smallFile, smallContent, 0644); err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}

	// Create a larger file
	largeFile := filepath.Join(tempDir, "large.txt")
	largeContent := make([]byte, 2048) // 2KB
	for i := range largeContent {
		largeContent[i] = byte('A')
	}
	if err := os.WriteFile(largeFile, largeContent, 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	tests := []struct {
		name                string
		maxFileSize         string
		filePath            string
		expectedWithinLimit bool
		expectedSize        int64
		expectError         bool
	}{
		{
			name:                "Small file within limit",
			maxFileSize:         "1KB",
			filePath:            smallFile,
			expectedWithinLimit: true,
			expectedSize:        int64(len(smallContent)),
			expectError:         false,
		},
		{
			name:                "Large file exceeds limit",
			maxFileSize:         "1KB",
			filePath:            largeFile,
			expectedWithinLimit: false,
			expectedSize:        int64(len(largeContent)),
			expectError:         false,
		},
		{
			name:                "File within large limit",
			maxFileSize:         "10KB",
			filePath:            largeFile,
			expectedWithinLimit: true,
			expectedSize:        int64(len(largeContent)),
			expectError:         false,
		},
		{
			name:                "No file size limit",
			maxFileSize:         "",
			filePath:            largeFile,
			expectedWithinLimit: true,
			expectedSize:        int64(len(largeContent)),
			expectError:         false,
		},
		{
			name:                "Non-existent file",
			maxFileSize:         "1KB",
			filePath:            "/non/existent/file",
			expectedWithinLimit: false,
			expectedSize:        0,
			expectError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter, err := NewSizeLimiter(tt.maxFileSize, 0)
			if err != nil {
				t.Fatalf("Failed to create size limiter: %v", err)
			}

			withinLimit, size, err := limiter.CheckFileSize(tt.filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for file %s, but got none", tt.filePath)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for file %s: %v", tt.filePath, err)
				return
			}

			if withinLimit != tt.expectedWithinLimit {
				t.Errorf("Expected withinLimit %v for file %s, got %v", tt.expectedWithinLimit, tt.filePath, withinLimit)
			}

			if size != tt.expectedSize {
				t.Errorf("Expected size %d for file %s, got %d", tt.expectedSize, tt.filePath, size)
			}
		})
	}
}

func TestSizeLimiter_AddToTotalSize(t *testing.T) {
	tests := []struct {
		name         string
		maxTotalSize int64
		additions    []int64
		expected     []bool
	}{
		{
			name:         "No limit",
			maxTotalSize: 0,
			additions:    []int64{100, 200, 300},
			expected:     []bool{true, true, true},
		},
		{
			name:         "Within limit",
			maxTotalSize: 1000,
			additions:    []int64{100, 200, 300},
			expected:     []bool{true, true, true},
		},
		{
			name:         "Exceeds limit",
			maxTotalSize: 500,
			additions:    []int64{200, 200, 200},
			expected:     []bool{true, true, false},
		},
		{
			name:         "Exactly at limit",
			maxTotalSize: 600,
			additions:    []int64{200, 200, 200},
			expected:     []bool{true, true, true},
		},
		{
			name:         "First addition exceeds",
			maxTotalSize: 100,
			additions:    []int64{200},
			expected:     []bool{false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter, err := NewSizeLimiter("1MB", tt.maxTotalSize)
			if err != nil {
				t.Fatalf("Failed to create size limiter: %v", err)
			}

			if len(tt.additions) != len(tt.expected) {
				t.Fatalf("Test setup error: additions and expected must have same length")
			}

			for i, addition := range tt.additions {
				result := limiter.AddToTotalSize(addition)
				if result != tt.expected[i] {
					t.Errorf("Addition %d: expected %v, got %v", i, tt.expected[i], result)
				}
			}

			// Check that CurrentTotalSize is updated correctly
			expectedTotal := int64(0)
			for _, addition := range tt.additions {
				expectedTotal += addition
			}

			if limiter.CurrentTotalSize != expectedTotal {
				t.Errorf("Expected CurrentTotalSize %d, got %d", expectedTotal, limiter.CurrentTotalSize)
			}
		})
	}
}

func TestSizeLimiter_GetTruncatedMessage(t *testing.T) {
	limiter, err := NewSizeLimiter("1MB", 10000)
	if err != nil {
		t.Fatalf("Failed to create size limiter: %v", err)
	}

	message := limiter.GetTruncatedMessage()
	expected := "[Output truncated: reached character limit of 10000]"

	if message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, message)
	}
}

func TestSizeLimiter_GetFileTooLargeMessage(t *testing.T) {
	limiter, err := NewSizeLimiter("1MB", 0)
	if err != nil {
		t.Fatalf("Failed to create size limiter: %v", err)
	}

	path := "/path/to/large/file.txt"
	size := int64(5 * 1024 * 1024) // 5MB

	message := limiter.GetFileTooLargeMessage(path, size)

	// Check that the message contains expected elements
	if message == "" {
		t.Error("Expected non-empty message")
	}

	// Should contain file size in MB
	if !containsSubstring(message, "5.0MB") {
		t.Errorf("Expected message to contain '5.0MB', got: %s", message)
	}

	// Should contain max size in MB
	if !containsSubstring(message, "1.0MB") {
		t.Errorf("Expected message to contain '1.0MB', got: %s", message)
	}

	// Should indicate file was skipped
	if !containsSubstring(message, "skipped") {
		t.Errorf("Expected message to contain 'skipped', got: %s", message)
	}
}

func TestSizeLimit_Struct(t *testing.T) {
	// Test SizeLimit structure
	limit := SizeLimit{MaxBytes: 1024}

	if limit.MaxBytes != 1024 {
		t.Errorf("Expected MaxBytes to be 1024, got %d", limit.MaxBytes)
	}
}

func TestSizeLimiter_ZeroLimits(t *testing.T) {
	// Test behavior with zero limits
	limiter, err := NewSizeLimiter("", 0)
	if err != nil {
		t.Fatalf("Failed to create size limiter: %v", err)
	}

	// Zero total limit should mean no limit
	result := limiter.AddToTotalSize(1000000)
	if !result {
		t.Error("Expected AddToTotalSize to return true when total limit is 0")
	}

	result = limiter.AddToTotalSize(1000000)
	if !result {
		t.Error("Expected AddToTotalSize to return true when total limit is 0")
	}
}

func TestSizeLimiter_Integration(t *testing.T) {
	// Integration test combining file size and total size limits
	tempDir, err := os.MkdirTemp("", "size_limiter_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	smallFile := filepath.Join(tempDir, "small.txt")
	if err := os.WriteFile(smallFile, []byte("small"), 0644); err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}

	largeFile := filepath.Join(tempDir, "large.txt")
	largeContent := make([]byte, 2048) // 2KB
	if err := os.WriteFile(largeFile, largeContent, 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	limiter, err := NewSizeLimiter("1KB", 8) // Set very small total limit
	if err != nil {
		t.Fatalf("Failed to create size limiter: %v", err)
	}

	// Check small file (should be within file size limit)
	withinLimit, size, err := limiter.CheckFileSize(smallFile)
	if err != nil {
		t.Fatalf("Failed to check small file size: %v", err)
	}
	if !withinLimit {
		t.Error("Expected small file to be within limit")
	}

	// Check large file (should exceed file size limit)
	withinLimit, _, err = limiter.CheckFileSize(largeFile)
	if err != nil {
		t.Fatalf("Failed to check large file size: %v", err)
	}
	if withinLimit {
		t.Error("Expected large file to exceed limit")
	}

	// Test total size limit
	result := limiter.AddToTotalSize(size) // Add small file size (5 bytes)
	if !result {
		t.Error("Expected first addition to be within total limit")
	}

	result = limiter.AddToTotalSize(size) // Add again (5+5=10 bytes, should exceed 8 byte limit)
	if result {
		t.Error("Expected second addition to exceed total limit")
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) != -1
}

// Simple substring search
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}