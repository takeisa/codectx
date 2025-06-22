package stats

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"codectx/internal/analysis"
	"codectx/internal/git"
	"codectx/internal/utils"
)

// AdvancedStatsCollector extends the basic StatsCollector with advanced statistics
type AdvancedStatsCollector struct {
	*StatsCollector
	rootDir            string
	HealthCheck        *analysis.HealthCheck
	ComplexityAnalysis *analysis.ComplexityAnalysis
	LanguageStats      *analysis.LanguageStats
	GitInfo            *git.GitInfo
	GitStatusSummary   *git.GitStatusSummary
}

// NewAdvancedStatsCollector creates a new advanced stats collector
func NewAdvancedStatsCollector() *AdvancedStatsCollector {
	return &AdvancedStatsCollector{
		StatsCollector: NewStatsCollector(),
	}
}

// CollectAdvancedStats collects advanced statistics for a directory
func CollectAdvancedStats(rootDir string, options AdvancedStatsOptions) (*AdvancedStatsCollector, error) {
	stats := NewAdvancedStatsCollector()
	stats.rootDir = rootDir

	// Collect basic stats
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

	// Collect advanced stats based on options
	if options.HealthCheck {
		healthCheck, err := analysis.CheckProjectHealth(rootDir, 10*1024*1024) // 10MB threshold for large files
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to check project health: %v\n", err)
		} else {
			stats.HealthCheck = healthCheck
		}
	}

	if options.ComplexityAnalysis {
		complexityAnalysis, err := analysis.AnalyzeProjectComplexity(rootDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to analyze project complexity: %v\n", err)
		} else {
			stats.ComplexityAnalysis = complexityAnalysis
		}
	}

	if options.LanguageStats {
		languageStats, err := analysis.AnalyzeLanguages(rootDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to analyze languages: %v\n", err)
		} else {
			stats.LanguageStats = languageStats
		}
	}

	if options.GitInfo {
		gitInfo, err := git.GetGitInfo(rootDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to get Git info: %v\n", err)
		} else {
			stats.GitInfo = gitInfo
		}
	}

	if options.GitStatus {
		gitStatusSummary, err := git.GetGitStatusSummary(rootDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to get Git status: %v\n", err)
		} else {
			stats.GitStatusSummary = gitStatusSummary
		}
	}

	return stats, nil
}

// PrintAdvancedStats prints the advanced statistics
func (s *AdvancedStatsCollector) PrintAdvancedStats() {
	// Print basic stats
	s.PrintStats()

	// Print health check if available
	if s.HealthCheck != nil {
		analysis.PrintHealthCheck(s.HealthCheck)
	}

	// Print complexity analysis if available
	if s.ComplexityAnalysis != nil {
		analysis.PrintComplexityAnalysis(s.ComplexityAnalysis)
	}

	// Print language stats if available
	if s.LanguageStats != nil {
		analysis.PrintLanguageStats(s.LanguageStats)
	}

	// Print Git status if available
	if s.GitStatusSummary != nil {
		fmt.Println("\nGit Status:")
		fmt.Println("===========")
		fmt.Printf("  Tracked files: %d/%d\n", s.GitStatusSummary.TrackedFiles, s.GitStatusSummary.TotalFiles)
		fmt.Printf("  Modified files: %d\n", s.GitStatusSummary.ModifiedFiles)
		fmt.Printf("  Untracked files: %d\n", s.GitStatusSummary.UntrackedFiles)
		fmt.Printf("  Last commit: %s\n", s.GitStatusSummary.LastCommitTime)
	}
}

// AdvancedStatsOptions defines options for collecting advanced statistics
type AdvancedStatsOptions struct {
	HealthCheck        bool
	ComplexityAnalysis bool
	LanguageStats      bool
	GitInfo            bool
	GitStatus          bool
}

// GetTopFileExtensions returns the top file extensions by count
func (s *AdvancedStatsCollector) GetTopFileExtensions(limit int) []ExtensionStat {
	// Count files by extension
	extCount := make(map[string]int)
	extSize := make(map[string]int64)

	filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != "" {
			extCount[ext]++
			extSize[ext] += info.Size()
		}
		return nil
	})

	// Convert to slice for sorting
	var stats []ExtensionStat
	for ext, count := range extCount {
		stats = append(stats, ExtensionStat{
			Extension: ext,
			Count:     count,
			Size:      extSize[ext],
		})
	}

	// Sort by count (descending)
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	// Limit results
	if limit > 0 && limit < len(stats) {
		stats = stats[:limit]
	}

	return stats
}

// ExtensionStat represents statistics for a file extension
type ExtensionStat struct {
	Extension string
	Count     int
	Size      int64
}

// GetAverageFileSize returns the average file size
func (s *AdvancedStatsCollector) GetAverageFileSize() float64 {
	if s.TotalFiles == 0 {
		return 0
	}
	return float64(s.TotalSize) / float64(s.TotalFiles)
}

// GetFileSizeDistribution returns the distribution of file sizes
func (s *AdvancedStatsCollector) GetFileSizeDistribution() map[string]int {
	distribution := map[string]int{
		"0-1KB":      0,
		"1KB-10KB":   0,
		"10KB-100KB": 0,
		"100KB-1MB":  0,
		"1MB-10MB":   0,
		"10MB+":      0,
	}

	filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		size := info.Size()
		switch {
		case size < 1024:
			distribution["0-1KB"]++
		case size < 10*1024:
			distribution["1KB-10KB"]++
		case size < 100*1024:
			distribution["10KB-100KB"]++
		case size < 1024*1024:
			distribution["100KB-1MB"]++
		case size < 10*1024*1024:
			distribution["1MB-10MB"]++
		default:
			distribution["10MB+"]++
		}
		return nil
	})

	return distribution
}

// GetModificationTimeStats returns statistics about file modification times
func (s *AdvancedStatsCollector) GetModificationTimeStats() ModTimeStats {
	var stats ModTimeStats
	now := time.Now()

	filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		modTime := info.ModTime()
		age := now.Sub(modTime)

		switch {
		case age < 24*time.Hour:
			stats.Last24Hours++
		case age < 7*24*time.Hour:
			stats.LastWeek++
		case age < 30*24*time.Hour:
			stats.LastMonth++
		case age < 365*24*time.Hour:
			stats.LastYear++
		default:
			stats.Older++
		}

		if modTime.Before(stats.OldestFile) || stats.OldestFile.IsZero() {
			stats.OldestFile = modTime
		}
		if modTime.After(stats.NewestFile) {
			stats.NewestFile = modTime
		}

		return nil
	})

	return stats
}

// ModTimeStats represents statistics about file modification times
type ModTimeStats struct {
	Last24Hours int
	LastWeek    int
	LastMonth   int
	LastYear    int
	Older       int
	OldestFile  time.Time
	NewestFile  time.Time
}
