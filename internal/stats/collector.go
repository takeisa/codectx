package stats

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"codectx/internal/utils"
)

// StatsCollector collects statistics about the scanned files
type StatsCollector struct {
	TotalFiles       int
	TotalDirectories int
	TotalSize        int64
	TextFiles        int
	BinaryFiles      int
	EstimatedTokens  int
	StartTime        time.Time
}

// NewStatsCollector creates a new stats collector
func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		StartTime: time.Now(),
	}
}

// AddFile adds a file to the statistics
func (s *StatsCollector) AddFile(path string, isText bool) error {
	// Get file info
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Update statistics
	s.TotalFiles++
	s.TotalSize += fileInfo.Size()

	if isText {
		s.TextFiles++
		// More accurate token estimation based on file content
		tokens, err := EstimateTokens(path)
		if err == nil {
			s.EstimatedTokens += tokens
		} else {
			// Fallback to rough estimate: 1 token per 4 bytes
			s.EstimatedTokens += int(fileInfo.Size() / 4)
		}
	} else {
		s.BinaryFiles++
	}

	return nil
}

// AddDirectory adds a directory to the statistics
func (s *StatsCollector) AddDirectory(path string) {
	s.TotalDirectories++
}

// GetProcessingTime returns the processing time in seconds
func (s *StatsCollector) GetProcessingTime() float64 {
	return time.Since(s.StartTime).Seconds()
}

// PrintStats prints the statistics
func (s *StatsCollector) PrintStats() {
	fmt.Println("\nStatistics:")
	fmt.Printf("  Total files: %d\n", s.TotalFiles)
	fmt.Printf("  Total directories: %d\n", s.TotalDirectories)
	fmt.Printf("  Total size: %.1fMB\n", float64(s.TotalSize)/(1024*1024))
	fmt.Printf("  Text files: %d\n", s.TextFiles)
	fmt.Printf("  Binary files: %d\n", s.BinaryFiles)
	fmt.Printf("  Estimated tokens: ~%d\n", s.EstimatedTokens)
	fmt.Printf("  Processing time: %.3fs\n", s.GetProcessingTime())
}

// CollectStats collects statistics for a directory
func CollectStats(rootDir string) (*StatsCollector, error) {
	stats := NewStatsCollector()

	// Walk the directory tree
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			stats.AddDirectory(path)
		} else {
			isText, err := utils.IsTextFile(path)
			if err != nil {
				return err
			}
			if err := stats.AddFile(path, isText); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to collect stats: %w", err)
	}

	return stats, nil
}


// EstimateTokens estimates the number of tokens in a text file
func EstimateTokens(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Get file extension for language-specific tokenization
	ext := strings.ToLower(filepath.Ext(path))

	var totalTokens int
	scanner := bufio.NewScanner(file)

	// Language-specific token estimation
	switch ext {
	case ".go", ".java", ".c", ".cpp", ".cc", ".cxx", ".h", ".hpp":
		// Code files: more tokens per word due to symbols
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
				continue // Skip empty lines and comments
			}
			totalTokens += estimateCodeLineTokens(line)
		}
	case ".js", ".ts", ".py", ".rb", ".php", ".cs", ".kt", ".swift", ".rs":
		// Script/interpreted languages
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
				continue
			}
			totalTokens += estimateCodeLineTokens(line)
		}
	case ".json", ".xml", ".yaml", ".yml", ".toml":
		// Structured data: fewer tokens per character
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			totalTokens += estimateDataLineTokens(line)
		}
	case ".md", ".txt", ".rst":
		// Text files: natural language tokenization
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			totalTokens += estimateTextLineTokens(line)
		}
	default:
		// Default estimation
		for scanner.Scan() {
			line := scanner.Text()
			totalTokens += len(line) / 4 // 1 token per 4 characters
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return totalTokens, nil
}

// estimateCodeLineTokens estimates tokens for a line of code
func estimateCodeLineTokens(line string) int {
	// Remove comments
	if idx := strings.Index(line, "//"); idx != -1 {
		line = line[:idx]
	}
	if idx := strings.Index(line, "#"); idx != -1 {
		line = line[:idx]
	}

	// Count words and symbols
	wordCount := len(strings.Fields(line))

	// Count programming symbols
	symbolRegex := regexp.MustCompile(`[{}()\[\];,.:+\-*/=<>!&|%^~]`)
	symbolCount := len(symbolRegex.FindAllString(line, -1))

	// Each word is ~1.3 tokens, each symbol is ~0.5 tokens in code
	return int(float64(wordCount)*1.3 + float64(symbolCount)*0.5)
}

// estimateDataLineTokens estimates tokens for structured data
func estimateDataLineTokens(line string) int {
	// Remove common data syntax
	cleaned := regexp.MustCompile(`[{}\[\]",:]`).ReplaceAllString(line, " ")
	words := strings.Fields(cleaned)

	// Data tokens are usually more efficient
	return int(float64(len(words)) * 1.1)
}

// estimateTextLineTokens estimates tokens for natural language text
func estimateTextLineTokens(line string) int {
	// Split by spaces and punctuation
	words := strings.FieldsFunc(line, func(c rune) bool {
		return unicode.IsSpace(c) || unicode.IsPunct(c)
	})

	// Filter out empty strings
	var realWords []string
	for _, word := range words {
		if strings.TrimSpace(word) != "" {
			realWords = append(realWords, word)
		}
	}

	// Natural language: ~1.3 tokens per word on average
	return int(float64(len(realWords)) * 1.3)
}
