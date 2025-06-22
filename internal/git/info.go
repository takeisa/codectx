package git

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GitInfo contains information about the Git repository
type GitInfo struct {
	CommitHash    string    `json:"commit_hash"`
	Branch        string    `json:"branch"`
	Author        string    `json:"author"`
	CommitDate    time.Time `json:"commit_date"`
	IsDirty       bool      `json:"is_dirty"`
	LastModified  time.Time `json:"last_modified"`
	RepositoryURL string    `json:"repository_url"`
}

// GetGitInfo retrieves Git information for the repository
func GetGitInfo(rootDir string) (*GitInfo, error) {
	// Check if git is available
	if !isGitCommandAvailable() {
		return nil, fmt.Errorf("git command not available")
	}

	// Check if the directory is a git repository
	if !isGitRepository(rootDir) {
		return nil, fmt.Errorf("not a git repository")
	}

	info := &GitInfo{}

	// Get commit hash
	hash, err := runGitCommand(rootDir, "rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get commit hash: %w", err)
	}
	info.CommitHash = strings.TrimSpace(hash)

	// Get branch
	branch, err := runGitCommand(rootDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}
	info.Branch = strings.TrimSpace(branch)

	// Get author
	author, err := runGitCommand(rootDir, "log", "-1", "--pretty=format:%an <%ae>")
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}
	info.Author = strings.TrimSpace(author)

	// Get commit date
	dateStr, err := runGitCommand(rootDir, "log", "-1", "--pretty=format:%aI")
	if err != nil {
		return nil, fmt.Errorf("failed to get commit date: %w", err)
	}
	date, err := time.Parse(time.RFC3339, strings.TrimSpace(dateStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse commit date: %w", err)
	}
	info.CommitDate = date

	// Check if the repository is dirty
	status, err := runGitCommand(rootDir, "status", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("failed to get git status: %w", err)
	}
	info.IsDirty = strings.TrimSpace(status) != ""

	// Get repository URL
	url, err := runGitCommand(rootDir, "config", "--get", "remote.origin.url")
	if err == nil {
		info.RepositoryURL = strings.TrimSpace(url)
	}

	// Get last modified time
	info.LastModified = time.Now()

	return info, nil
}

// GetGitTrackedFiles returns a list of files tracked by Git
func GetGitTrackedFiles(rootDir string) ([]string, error) {
	// Check if git is available
	if !isGitCommandAvailable() {
		return nil, fmt.Errorf("git command not available")
	}

	// Check if the directory is a git repository
	if !isGitRepository(rootDir) {
		return nil, fmt.Errorf("not a git repository")
	}

	// Get tracked files
	output, err := runGitCommand(rootDir, "ls-files")
	if err != nil {
		return nil, fmt.Errorf("failed to get tracked files: %w", err)
	}

	files := strings.Split(strings.TrimSpace(output), "\n")
	return filterEmptyStrings(files), nil
}

// GetGitStatus returns the status of files in the repository
func GetGitStatus(rootDir string) (map[string]string, error) {
	// Check if git is available
	if !isGitCommandAvailable() {
		return nil, fmt.Errorf("git command not available")
	}

	// Check if the directory is a git repository
	if !isGitRepository(rootDir) {
		return nil, fmt.Errorf("not a git repository")
	}

	// Get status
	output, err := runGitCommand(rootDir, "status", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("failed to get git status: %w", err)
	}

	status := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Format is XY PATH
		// X is status in staging area, Y is status in working tree
		statusCode := strings.TrimSpace(line[:2])
		path := strings.TrimSpace(line[3:])
		status[path] = statusCode
	}

	return status, nil
}

// Helper functions

// isGitCommandAvailable checks if the git command is available
func isGitCommandAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// isGitRepository checks if the directory is a git repository
func isGitRepository(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

// runGitCommand runs a git command and returns its output
func runGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// filterEmptyStrings removes empty strings from a slice
func filterEmptyStrings(slice []string) []string {
	var result []string
	for _, s := range slice {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}
