package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"codectx/internal/filter"
	"codectx/internal/formatter"
	"codectx/internal/git"
	"codectx/internal/limits"
	"codectx/internal/scanner"
	"codectx/internal/stats"
	"codectx/internal/utils"
)

// Command line flags
var (
	// Output format
	formatFlag string

	// Filtering options
	extensionsFlag  string
	excludeFlag     string
	includeDotfiles bool

	// Size limits
	limitFlag       int64
	maxFileSizeFlag string

	// Statistics
	statsFlag bool

	// Git integration
	gitOnlyFlag          bool
	respectGitignoreFlag bool
	ignoreGitignoreFlag  bool
	includeGitInfoFlag   bool
	gitStatusFlag        bool

	// Advanced analysis
	healthCheckFlag        bool
	complexityAnalysisFlag bool
	languageStatsFlag      bool

	// Other options
	outputFlag        string
	noLineNumbersFlag bool
	verboseFlag       bool
	helpFlag          bool
	versionFlag       bool
	dryRunFlag        bool
)

// Execute runs the root command
func Execute() error {
	// Define flags
	flag.StringVar(&formatFlag, "format", "text", "Output format (text, html, markdown, json)")
	flag.StringVar(&formatFlag, "f", "text", "Output format (short)")

	flag.StringVar(&extensionsFlag, "extensions", "", "Filter by file extensions (comma-separated)")
	flag.StringVar(&extensionsFlag, "e", "", "Filter by file extensions (short)")

	flag.StringVar(&excludeFlag, "exclude", "", "Exclude patterns (comma-separated)")
	flag.StringVar(&excludeFlag, "x", "", "Exclude patterns (short)")

	flag.BoolVar(&includeDotfiles, "include-dotfiles", false, "Include dotfiles")

	flag.Int64Var(&limitFlag, "limit", 0, "Maximum total character limit (0 for no limit)")
	flag.Int64Var(&limitFlag, "l", 0, "Maximum total character limit (short)")

	flag.StringVar(&maxFileSizeFlag, "max-file-size", "1MB", "Maximum file size (e.g., 1MB, 500KB)")

	flag.BoolVar(&statsFlag, "stats", false, "Show statistics")

	flag.StringVar(&outputFlag, "output", "", "Output file")
	flag.StringVar(&outputFlag, "o", "", "Output file (short)")

	flag.BoolVar(&noLineNumbersFlag, "no-line-numbers", false, "Don't show line numbers")
	flag.BoolVar(&noLineNumbersFlag, "n", false, "Don't show line numbers (short)")

	flag.BoolVar(&verboseFlag, "verbose", false, "Verbose output")
	flag.BoolVar(&verboseFlag, "v", false, "Verbose output (short)")

	flag.BoolVar(&helpFlag, "help", false, "Show help")
	flag.BoolVar(&helpFlag, "h", false, "Show help (short)")

	flag.BoolVar(&versionFlag, "version", false, "Show version")

	flag.BoolVar(&dryRunFlag, "dry-run", false, "Show files that would be processed without processing them")

	// Git integration flags
	flag.BoolVar(&gitOnlyFlag, "git-only", false, "Only include Git tracked files")
	flag.BoolVar(&respectGitignoreFlag, "respect-gitignore", false, "Respect .gitignore patterns")
	flag.BoolVar(&ignoreGitignoreFlag, "ignore-gitignore", true, "Ignore .gitignore patterns (default)")
	flag.BoolVar(&includeGitInfoFlag, "include-git-info", false, "Include Git information in output")
	flag.BoolVar(&gitStatusFlag, "git-status", false, "Show Git status information")

	// Advanced analysis flags
	flag.BoolVar(&healthCheckFlag, "health-check", false, "Perform project health check")
	flag.BoolVar(&complexityAnalysisFlag, "complexity-analysis", false, "Perform complexity analysis")
	flag.BoolVar(&languageStatsFlag, "language-stats", false, "Show language statistics")

	// Parse flags
	flag.Parse()

	// Show help
	if helpFlag {
		printHelp()
		return nil
	}

	// Show version
	if versionFlag {
		fmt.Println("codectx v0.1.0")
		return nil
	}

	// Get target directory
	targetDir := "."
	args := flag.Args()
	if len(args) > 0 {
		targetDir = args[0]
	}

	// Resolve absolute path
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Check if directory exists
	info, err := os.Stat(absTargetDir)
	if err != nil {
		return fmt.Errorf("failed to access target directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", absTargetDir)
	}

	// Run the command
	return run(absTargetDir)
}

// run executes the main functionality
func run(targetDir string) error {
	if verboseFlag {
		fmt.Printf("Scanning directory: %s\n", targetDir)
	}

	// Initialize stats collector if stats flag is set
	var statsCollector *stats.StatsCollector
	var advancedStatsCollector *stats.AdvancedStatsCollector

	// Check if any advanced stats options are enabled
	advancedStatsEnabled := statsFlag && (healthCheckFlag || complexityAnalysisFlag || languageStatsFlag)

	if advancedStatsEnabled {
		// Use advanced stats collector
		options := stats.AdvancedStatsOptions{
			HealthCheck:        healthCheckFlag,
			ComplexityAnalysis: complexityAnalysisFlag,
			LanguageStats:      languageStatsFlag,
			GitInfo:            includeGitInfoFlag,
			GitStatus:          gitStatusFlag,
		}

		var err error
		advancedStatsCollector, err = stats.CollectAdvancedStats(targetDir, options)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to collect advanced stats: %v\n", err)
		}

		// Use the basic stats collector from the advanced one
		statsCollector = advancedStatsCollector.StatsCollector
	} else if statsFlag {
		// Use basic stats collector
		statsCollector = stats.NewStatsCollector()
	}

	// Handle Git status flag
	if gitStatusFlag {
		if err := git.PrintGitStatus(targetDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to get Git status: %v\n", err)
		}
		// If only Git status is requested, return after printing it
		if !statsFlag && !gitOnlyFlag && !includeGitInfoFlag {
			return nil
		}
	}

	// Get Git tracked files if --git-only is specified
	var gitTrackedFiles []string
	if gitOnlyFlag {
		var err error
		gitTrackedFiles, err = git.GetGitTrackedFiles(targetDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to get Git tracked files: %v\n", err)
			fmt.Fprintf(os.Stderr, "Continuing without Git tracking filter\n")
		}
	}

	// Get Git info if --include-git-info is specified
	var gitInfo *git.GitInfo
	if includeGitInfoFlag {
		var err error
		gitInfo, err = git.GetGitInfo(targetDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to get Git info: %v\n", err)
		}
	}

	// Create a scanner
	scanner := scanner.NewScanner(targetDir, includeDotfiles)

	// Scan the directory
	root, err := scanner.Scan()
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	// Generate the tree
	tree := scanner.GenerateTree(root)

	// Create a filter
	filter := filter.NewFilter(extensionsFlag, excludeFlag, includeDotfiles)

	// Handle .gitignore if needed
	if respectGitignoreFlag && !ignoreGitignoreFlag {
		gitIgnoreParser := git.NewGitIgnoreParser(targetDir)
		if err := gitIgnoreParser.ParseAllGitIgnores(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse .gitignore files: %v\n", err)
		} else {
			filter.SetGitIgnoreParser(gitIgnoreParser)
		}
	}

	// Set Git tracked files if --git-only is specified
	if gitOnlyFlag && len(gitTrackedFiles) > 0 {
		filter.SetGitTrackedFiles(gitTrackedFiles)
	}

	// Create a size limiter
	sizeLimiter, err := limits.NewSizeLimiter(maxFileSizeFlag, limitFlag)
	if err != nil {
		return fmt.Errorf("failed to create size limiter: %w", err)
	}

	// Create a formatter
	formatter, err := formatter.NewFormatter(formatFlag, !noLineNumbersFlag, outputFlag, sizeLimiter, gitInfo)
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}
	defer formatter.Close()

	// Format the tree
	if err := formatter.FormatTree(tree); err != nil {
		return fmt.Errorf("failed to format tree: %w", err)
	}

	// Get all file paths
	paths := scanner.GetRelativePaths(root)

	// Count directories for stats
	if statsCollector != nil {
		// Count the root directory
		statsCollector.AddDirectory(targetDir)

		// Count all subdirectories
		for _, child := range root.Children {
			if child.IsDir {
				countDirectories(child, statsCollector)
			}
		}
	}

	// Process each file
	for _, relPath := range paths {
		fullPath := filepath.Join(targetDir, relPath[1:]) // Remove leading slash
		cleanRelPath := relPath[1:] // Clean relative path without leading slash

		// Check if the file should be included
		if !filter.ShouldInclude(fullPath) {
			if verboseFlag {
				fmt.Fprintf(os.Stderr, "Skipping file: %s\n", cleanRelPath)
			}
			continue
		}

		// Check if it's a text file
		isText, err := utils.IsTextFile(fullPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to check if file is text: %v\n", err)
			continue
		}

		if !isText {
			fmt.Fprintf(os.Stderr, "Warning: skipping binary file: %s\n", cleanRelPath)
			continue
		}

		// Update stats if stats flag is set
		if statsCollector != nil {
			if err := statsCollector.AddFile(fullPath, isText); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to add file to stats: %v\n", err)
			}
		}

		// If dry run flag is set, just print the file path and skip formatting
		if dryRunFlag {
			fmt.Fprintf(os.Stderr, "Would process file: %s\n", cleanRelPath)
			continue
		}

		// Format the file content
		if err := formatter.FormatFileContent(fullPath, cleanRelPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to format file content: %v\n", err)
			continue
		}
	}

	// Print stats if stats flag is set
	if advancedStatsCollector != nil {
		advancedStatsCollector.PrintAdvancedStats()
	} else if statsCollector != nil {
		statsCollector.PrintStats()
	}

	return nil
}

// countDirectories recursively counts directories
func countDirectories(entry *scanner.FileEntry, statsCollector *stats.StatsCollector) {
	if entry.IsDir {
		statsCollector.AddDirectory(entry.Path)
		for _, child := range entry.Children {
			if child.IsDir {
				countDirectories(child, statsCollector)
			}
		}
	}
}

// printHelp shows the help message
func printHelp() {
	fmt.Println("codectx - Unified directory and file content viewer")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  codectx [TARGET_DIR] [OPTIONS]")
	fmt.Println("")
	fmt.Println("Arguments:")
	fmt.Println("  TARGET_DIR    Directory to scan (default: current directory)")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -f, --format <FORMAT>                Output format (text, html, markdown, json)")
	fmt.Println("  -e, --extensions <EXT1,EXT2,...>     Filter by file extensions")
	fmt.Println("  -x, --exclude <PATTERN1,PATTERN2,..> Exclude patterns")
	fmt.Println("      --include-dotfiles               Include dotfiles")
	fmt.Println("  -l, --limit <NUMBER>                 Maximum total character limit (0 for no limit)")
	fmt.Println("      --max-file-size <SIZE>           Maximum file size (e.g., 1MB, 500KB)")
	fmt.Println("      --stats                          Show statistics")
	fmt.Println("  -o, --output <FILE>                  Output file (default: stdout)")
	fmt.Println("  -n, --no-line-numbers                Don't show line numbers")
	fmt.Println("  -v, --verbose                        Verbose output")
	fmt.Println("  -h, --help                           Show help")
	fmt.Println("      --version                        Show version")
	fmt.Println("      --dry-run                        Show files without processing")
	fmt.Println("")
	fmt.Println("Git Integration Options:")
	fmt.Println("      --git-only                       Only include Git tracked files")
	fmt.Println("      --respect-gitignore              Respect .gitignore patterns")
	fmt.Println("      --ignore-gitignore               Ignore .gitignore patterns (default)")
	fmt.Println("      --include-git-info               Include Git information in output")
	fmt.Println("      --git-status                     Show Git status information")
	fmt.Println("")
	fmt.Println("Advanced Analysis Options:")
	fmt.Println("      --health-check                   Perform project health check")
	fmt.Println("      --complexity-analysis            Perform complexity analysis")
	fmt.Println("      --language-stats                 Show language statistics")
}
