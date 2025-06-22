package analysis

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ComplexityAnalysis represents the complexity analysis results for a project
type ComplexityAnalysis struct {
	TotalLines      int                `json:"total_lines"`
	CodeLines       int                `json:"code_lines"`
	CommentLines    int                `json:"comment_lines"`
	BlankLines      int                `json:"blank_lines"`
	CodeDensity     float64            `json:"code_density"`
	ComplexFiles    []ComplexFileInfo  `json:"complex_files"`
	LanguageMetrics map[string]Metrics `json:"language_metrics"`
}

// ComplexFileInfo contains complexity information about a file
type ComplexFileInfo struct {
	Path            string  `json:"path"`
	Lines           int     `json:"lines"`
	ComplexityScore float64 `json:"complexity_score"`
}

// Metrics contains metrics for a specific language
type Metrics struct {
	Files      int     `json:"files"`
	Lines      int     `json:"lines"`
	CodeLines  int     `json:"code_lines"`
	BlankLines int     `json:"blank_lines"`
	Comments   int     `json:"comments"`
	Percentage float64 `json:"percentage"`
}

// NewComplexityAnalysis creates a new complexity analysis
func NewComplexityAnalysis() *ComplexityAnalysis {
	return &ComplexityAnalysis{
		ComplexFiles:    []ComplexFileInfo{},
		LanguageMetrics: make(map[string]Metrics),
	}
}

// AnalyzeProjectComplexity performs a complexity analysis on the project
func AnalyzeProjectComplexity(rootDir string) (*ComplexityAnalysis, error) {
	analysis := NewComplexityAnalysis()

	// Walk the directory tree to analyze files
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if info.IsDir() && filepath.Base(path) == ".git" {
			return filepath.SkipDir
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get file extension
		ext := strings.ToLower(filepath.Ext(path))
		if ext == "" {
			return nil
		}

		// Remove the leading dot
		ext = ext[1:]

		// Analyze file complexity
		fileMetrics, err := analyzeFileComplexity(path, ext)
		if err != nil {
			return nil
		}

		// Update total metrics
		analysis.TotalLines += fileMetrics.Lines
		analysis.CodeLines += fileMetrics.CodeLines
		analysis.CommentLines += fileMetrics.Comments
		analysis.BlankLines += fileMetrics.BlankLines

		// Update language metrics
		if metrics, ok := analysis.LanguageMetrics[ext]; ok {
			metrics.Files++
			metrics.Lines += fileMetrics.Lines
			metrics.CodeLines += fileMetrics.CodeLines
			metrics.BlankLines += fileMetrics.BlankLines
			metrics.Comments += fileMetrics.Comments
			analysis.LanguageMetrics[ext] = metrics
		} else {
			analysis.LanguageMetrics[ext] = Metrics{
				Files:      1,
				Lines:      fileMetrics.Lines,
				CodeLines:  fileMetrics.CodeLines,
				BlankLines: fileMetrics.BlankLines,
				Comments:   fileMetrics.Comments,
			}
		}

		// Add complex files
		if fileMetrics.Lines > 300 || fileMetrics.ComplexityScore > 20 {
			relPath, err := filepath.Rel(rootDir, path)
			if err == nil {
				analysis.ComplexFiles = append(analysis.ComplexFiles, ComplexFileInfo{
					Path:            relPath,
					Lines:           fileMetrics.Lines,
					ComplexityScore: fileMetrics.ComplexityScore,
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to analyze project complexity: %w", err)
	}

	// Calculate code density
	if analysis.TotalLines > 0 {
		analysis.CodeDensity = float64(analysis.CodeLines) / float64(analysis.TotalLines) * 100
	}

	// Calculate language percentages
	totalFiles := 0
	for _, metrics := range analysis.LanguageMetrics {
		totalFiles += metrics.Files
	}

	if totalFiles > 0 {
		for lang, metrics := range analysis.LanguageMetrics {
			metrics.Percentage = float64(metrics.Files) / float64(totalFiles) * 100
			analysis.LanguageMetrics[lang] = metrics
		}
	}

	return analysis, nil
}

// PrintComplexityAnalysis prints the complexity analysis results
func PrintComplexityAnalysis(analysis *ComplexityAnalysis) {
	fmt.Println("\nComplexity Analysis:")
	fmt.Println("===================")

	fmt.Println("\nCode Metrics:")
	fmt.Printf("  Total lines: %d\n", analysis.TotalLines)
	fmt.Printf("  Code lines: %d (%.1f%%)\n", analysis.CodeLines, float64(analysis.CodeLines)/float64(analysis.TotalLines)*100)
	fmt.Printf("  Comment lines: %d (%.1f%%)\n", analysis.CommentLines, float64(analysis.CommentLines)/float64(analysis.TotalLines)*100)
	fmt.Printf("  Blank lines: %d (%.1f%%)\n", analysis.BlankLines, float64(analysis.BlankLines)/float64(analysis.TotalLines)*100)
	fmt.Printf("  Code density: %.1f%%\n", analysis.CodeDensity)

	// Print language metrics
	if len(analysis.LanguageMetrics) > 0 {
		fmt.Println("\nLanguage Distribution:")
		for lang, metrics := range analysis.LanguageMetrics {
			fmt.Printf("  %s: %d files (%.1f%%) - %d lines\n",
				lang, metrics.Files, metrics.Percentage, metrics.Lines)
		}
	}

	// Print complex files
	if len(analysis.ComplexFiles) > 0 {
		fmt.Println("\nComplex Files:")
		for _, file := range analysis.ComplexFiles {
			fmt.Printf("  %s: %d lines, complexity score: %.1f\n",
				file.Path, file.Lines, file.ComplexityScore)
		}
	}
}

// Helper functions

// FileMetrics contains metrics for a specific file
type FileMetrics struct {
	Lines           int
	CodeLines       int
	BlankLines      int
	Comments        int
	ComplexityScore float64
}

// analyzeFileComplexity analyzes the complexity of a file
func analyzeFileComplexity(path, ext string) (*FileMetrics, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	metrics := &FileMetrics{}

	// Define comment patterns based on file extension
	var lineCommentPattern, blockCommentStartPattern, blockCommentEndPattern string

	switch ext {
	case "go", "c", "cpp", "java", "js", "ts", "cs", "php", "swift":
		lineCommentPattern = "//"
		blockCommentStartPattern = "/*"
		blockCommentEndPattern = "*/"
	case "py", "rb", "sh", "bash":
		lineCommentPattern = "#"
	case "sql":
		lineCommentPattern = "--"
		blockCommentStartPattern = "/*"
		blockCommentEndPattern = "*/"
	case "html", "xml":
		blockCommentStartPattern = "<!--"
		blockCommentEndPattern = "-->"
	}

	// Complexity indicators
	nestedControlStructurePattern := regexp.MustCompile(`\s+(if|for|while|switch)\s+.*{`)
	functionPattern := regexp.MustCompile(`\s*func\s+\w+\s*\(`)

	inBlockComment := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		metrics.Lines++

		trimmedLine := strings.TrimSpace(line)

		// Check for blank lines
		if trimmedLine == "" {
			metrics.BlankLines++
			continue
		}

		// Check for block comments
		if blockCommentStartPattern != "" && blockCommentEndPattern != "" {
			if inBlockComment {
				metrics.Comments++
				if strings.Contains(line, blockCommentEndPattern) {
					inBlockComment = false
				}
				continue
			} else if strings.Contains(line, blockCommentStartPattern) {
				metrics.Comments++
				if !strings.Contains(line, blockCommentEndPattern) {
					inBlockComment = true
				}
				continue
			}
		}

		// Check for line comments
		if lineCommentPattern != "" && strings.HasPrefix(trimmedLine, lineCommentPattern) {
			metrics.Comments++
			continue
		}

		// Count code lines
		metrics.CodeLines++

		// Calculate complexity score
		if nestedControlStructurePattern.MatchString(line) {
			metrics.ComplexityScore += 1.0
		}

		if functionPattern.MatchString(line) {
			metrics.ComplexityScore += 0.5
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}
