package limits

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SizeLimit represents a size limit in bytes
type SizeLimit struct {
	MaxBytes int64
}

// SizeLimiter handles size limits for files and overall output
type SizeLimiter struct {
	MaxFileSize      int64 // Maximum size of individual files in bytes
	MaxTotalSize     int64 // Maximum total size of all output in bytes
	CurrentTotalSize int64 // Current total size of all output in bytes
}

// NewSizeLimiter creates a new size limiter with the given limits
func NewSizeLimiter(maxFileSize string, maxTotalSize int64) (*SizeLimiter, error) {
	var maxFileSizeBytes int64
	if maxFileSize != "" {
		size, err := ParseSize(maxFileSize)
		if err != nil {
			return nil, err
		}
		maxFileSizeBytes = size
	} else {
		// Default to 1MB if not specified
		maxFileSizeBytes = 1024 * 1024
	}

	return &SizeLimiter{
		MaxFileSize:      maxFileSizeBytes,
		MaxTotalSize:     maxTotalSize,
		CurrentTotalSize: 0,
	}, nil
}

// CheckFileSize checks if a file exceeds the maximum file size
func (l *SizeLimiter) CheckFileSize(path string) (bool, int64, error) {
	// Get file info
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, 0, fmt.Errorf("failed to get file info: %w", err)
	}

	// Check if the file size exceeds the limit
	fileSize := fileInfo.Size()
	if l.MaxFileSize > 0 && fileSize > l.MaxFileSize {
		return false, fileSize, nil
	}

	return true, fileSize, nil
}

// AddToTotalSize adds the given size to the current total size and checks if it exceeds the limit
func (l *SizeLimiter) AddToTotalSize(size int64) bool {
	l.CurrentTotalSize += size
	return l.MaxTotalSize <= 0 || l.CurrentTotalSize <= l.MaxTotalSize
}

// GetTruncatedMessage returns a message indicating that output was truncated
func (l *SizeLimiter) GetTruncatedMessage() string {
	return fmt.Sprintf("[Output truncated: reached character limit of %d]", l.MaxTotalSize)
}

// GetFileTooLargeMessage returns a message indicating that a file was too large
func (l *SizeLimiter) GetFileTooLargeMessage(path string, size int64) string {
	return fmt.Sprintf("[File too large: %.1fMB - skipped (max: %.1fMB)]", 
		float64(size)/(1024*1024), float64(l.MaxFileSize)/(1024*1024))
}

// ParseSize parses a size string (e.g., "1MB", "500KB") into bytes
func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0, nil
	}

	// Find the last digit position
	lastDigitPos := -1
	for i, r := range sizeStr {
		if r < '0' || r > '9' {
			lastDigitPos = i
			break
		}
	}

	// If no unit specified, assume bytes
	if lastDigitPos == -1 {
		return strconv.ParseInt(sizeStr, 10, 64)
	}

	// Parse the numeric part
	valueStr := sizeStr[:lastDigitPos]
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size value: %s", valueStr)
	}

	// Parse the unit part
	unit := strings.ToUpper(strings.TrimSpace(sizeStr[lastDigitPos:]))
	switch unit {
	case "B":
		return value, nil
	case "KB":
		return value * 1024, nil
	case "MB":
		return value * 1024 * 1024, nil
	case "GB":
		return value * 1024 * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unknown size unit: %s", unit)
	}
}