package analysis

import (
	"fmt"
	"os"
	"path/filepath"
)

// HealthCheck represents the health check results for a project
type HealthCheck struct {
	HasReadme        bool     `json:"has_readme"`
	HasLicense       bool     `json:"has_license"`
	HasGitignore     bool     `json:"has_gitignore"`
	HasTests         bool     `json:"has_tests"`
	LargeFiles       []string `json:"large_files"`
	EmptyDirectories []string `json:"empty_directories"`
	BinaryFiles      int      `json:"binary_files_count"`
	Warnings         []string `json:"warnings"`
}

// NewHealthCheck creates a new health check
func NewHealthCheck() *HealthCheck {
	return &HealthCheck{
		LargeFiles:       []string{},
		EmptyDirectories: []string{},
		Warnings:         []string{},
	}
}

// CheckProjectHealth performs a health check on the project
func CheckProjectHealth(rootDir string, largeFileSizeThreshold int64) (*HealthCheck, error) {
	health := NewHealthCheck()

	// Check for important files
	health.HasReadme = fileExists(filepath.Join(rootDir, "README.md")) || fileExists(filepath.Join(rootDir, "readme.md"))
	health.HasLicense = fileExists(filepath.Join(rootDir, "LICENSE")) || fileExists(filepath.Join(rootDir, "license"))
	health.HasGitignore = fileExists(filepath.Join(rootDir, ".gitignore"))

	// Check for tests
	health.HasTests = directoryExists(filepath.Join(rootDir, "tests")) ||
		directoryExists(filepath.Join(rootDir, "test")) ||
		hasFilesWithSuffix(rootDir, "_test.go")

	// Walk the directory tree to find large files, empty directories, and binary files
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if info.IsDir() && filepath.Base(path) == ".git" {
			return filepath.SkipDir
		}

		// Check for empty directories
		if info.IsDir() && path != rootDir {
			empty, err := isEmptyDir(path)
			if err != nil {
				return err
			}
			if empty {
				relPath, err := filepath.Rel(rootDir, path)
				if err == nil {
					health.EmptyDirectories = append(health.EmptyDirectories, relPath)
				}
			}
		}

		// Check for large files
		if !info.IsDir() && info.Size() > largeFileSizeThreshold {
			relPath, err := filepath.Rel(rootDir, path)
			if err == nil {
				health.LargeFiles = append(health.LargeFiles, fmt.Sprintf("%s (%.2fMB)", relPath, float64(info.Size())/(1024*1024)))
			}
		}

		// Check for binary files
		if !info.IsDir() {
			isBinary, err := isBinaryFile(path)
			if err == nil && isBinary {
				health.BinaryFiles++
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to check project health: %w", err)
	}

	// Generate warnings
	if !health.HasReadme {
		health.Warnings = append(health.Warnings, "No README.md file found")
	}
	if !health.HasLicense {
		health.Warnings = append(health.Warnings, "No LICENSE file found")
	}
	if !health.HasGitignore {
		health.Warnings = append(health.Warnings, "No .gitignore file found")
	}
	if !health.HasTests {
		health.Warnings = append(health.Warnings, "No tests found")
	}
	if len(health.LargeFiles) > 0 {
		health.Warnings = append(health.Warnings, fmt.Sprintf("Large files detected: %d", len(health.LargeFiles)))
	}
	if len(health.EmptyDirectories) > 0 {
		health.Warnings = append(health.Warnings, fmt.Sprintf("Empty directories: %d", len(health.EmptyDirectories)))
	}
	if health.BinaryFiles > 0 {
		health.Warnings = append(health.Warnings, fmt.Sprintf("Binary files: %d (consider adding to .gitignore)", health.BinaryFiles))
	}

	return health, nil
}

// PrintHealthCheck prints the health check results
func PrintHealthCheck(health *HealthCheck) {
	fmt.Println("\nProject Health Check:")
	fmt.Println("=====================")

	// Print warnings
	if len(health.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, warning := range health.Warnings {
			fmt.Printf("⚠️  %s\n", warning)
		}
	}

	// Print positive checks
	fmt.Println("\nChecks:")
	printCheck(health.HasReadme, "README.md present")
	printCheck(health.HasLicense, "LICENSE file present")
	printCheck(health.HasGitignore, ".gitignore configured")
	printCheck(health.HasTests, "Tests present")

	// Print large files
	if len(health.LargeFiles) > 0 {
		fmt.Println("\nLarge files:")
		for _, file := range health.LargeFiles {
			fmt.Printf("  %s\n", file)
		}
	}

	// Print empty directories
	if len(health.EmptyDirectories) > 0 {
		fmt.Println("\nEmpty directories:")
		for _, dir := range health.EmptyDirectories {
			fmt.Printf("  %s\n", dir)
		}
	}
}

// Helper functions

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// directoryExists checks if a directory exists
func directoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// isEmptyDir checks if a directory is empty
func isEmptyDir(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == nil {
		// Directory is not empty
		return false, nil
	}
	if err.Error() == "EOF" {
		// Directory is empty
		return true, nil
	}
	return false, err
}

// isBinaryFile checks if a file is binary
func isBinaryFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read the first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && n == 0 {
		return false, err
	}

	// Check for null bytes, which indicate a binary file
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, nil
		}
	}

	return false, nil
}

// hasFilesWithSuffix checks if a directory has files with a specific suffix
func hasFilesWithSuffix(rootDir, suffix string) bool {
	found := false
	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == suffix {
			found = true
			return filepath.SkipDir
		}
		return nil
	})
	return found
}

// printCheck prints a check result
func printCheck(condition bool, message string) {
	if condition {
		fmt.Printf("✅ %s\n", message)
	} else {
		fmt.Printf("❌ %s\n", message)
	}
}
