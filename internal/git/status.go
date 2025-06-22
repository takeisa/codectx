package git

import (
	"fmt"
	"strings"
)

// FileStatus represents the Git status of a file
type FileStatus struct {
	Path       string `json:"path"`
	StatusCode string `json:"status_code"`
	Status     string `json:"status"`
	Tracked    bool   `json:"tracked"`
	Modified   bool   `json:"modified"`
	Staged     bool   `json:"staged"`
}

// GitStatusSummary provides a summary of the Git status
type GitStatusSummary struct {
	TotalFiles     int           `json:"total_files"`
	TrackedFiles   int           `json:"tracked_files"`
	UntrackedFiles int           `json:"untracked_files"`
	ModifiedFiles  int           `json:"modified_files"`
	StagedFiles    int           `json:"staged_files"`
	FileStatuses   []*FileStatus `json:"file_statuses"`
	BranchName     string        `json:"branch_name"`
	CommitHash     string        `json:"commit_hash"`
	LastCommitTime string        `json:"last_commit_time"`
	IsDirty        bool          `json:"is_dirty"`
}

// GetGitStatusSummary returns a summary of the Git status
func GetGitStatusSummary(rootDir string) (*GitStatusSummary, error) {
	// Check if git is available
	if !isGitCommandAvailable() {
		return nil, fmt.Errorf("git command not available")
	}

	// Check if the directory is a git repository
	if !isGitRepository(rootDir) {
		return nil, fmt.Errorf("not a git repository")
	}

	summary := &GitStatusSummary{
		FileStatuses: []*FileStatus{},
	}

	// Get Git info
	gitInfo, err := GetGitInfo(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get git info: %w", err)
	}

	summary.BranchName = gitInfo.Branch
	summary.CommitHash = gitInfo.CommitHash
	summary.LastCommitTime = gitInfo.CommitDate.Format("2006-01-02 15:04:05")
	summary.IsDirty = gitInfo.IsDirty

	// Get tracked files
	trackedFiles, err := GetGitTrackedFiles(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracked files: %w", err)
	}
	summary.TrackedFiles = len(trackedFiles)

	// Get file statuses
	statuses, err := GetGitStatus(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get git status: %w", err)
	}

	// Process tracked files
	for _, path := range trackedFiles {
		status := &FileStatus{
			Path:    path,
			Tracked: true,
		}

		if statusCode, ok := statuses[path]; ok {
			status.StatusCode = statusCode
			status.Status = getStatusDescription(statusCode)
			status.Modified = isModified(statusCode)
			status.Staged = isStaged(statusCode)

			if status.Modified {
				summary.ModifiedFiles++
			}
			if status.Staged {
				summary.StagedFiles++
			}
		} else {
			status.StatusCode = ""
			status.Status = "Unchanged"
		}

		summary.FileStatuses = append(summary.FileStatuses, status)
	}

	// Process untracked files
	for path, statusCode := range statuses {
		if !isTracked(trackedFiles, path) {
			status := &FileStatus{
				Path:       path,
				StatusCode: statusCode,
				Status:     getStatusDescription(statusCode),
				Tracked:    false,
				Modified:   false,
				Staged:     isStaged(statusCode),
			}

			if status.Staged {
				summary.StagedFiles++
			}

			summary.FileStatuses = append(summary.FileStatuses, status)
			summary.UntrackedFiles++
		}
	}

	summary.TotalFiles = summary.TrackedFiles + summary.UntrackedFiles

	return summary, nil
}

// PrintGitStatus prints the Git status in a human-readable format
func PrintGitStatus(rootDir string) error {
	summary, err := GetGitStatusSummary(rootDir)
	if err != nil {
		return err
	}

	fmt.Println("Git Status Summary:")
	fmt.Printf("  Branch: %s\n", summary.BranchName)
	fmt.Printf("  Commit: %s\n", summary.CommitHash)
	fmt.Printf("  Last commit: %s\n", summary.LastCommitTime)
	fmt.Printf("  Repository state: %s\n", getRepositoryState(summary.IsDirty))
	fmt.Println()
	fmt.Printf("  Total files: %d\n", summary.TotalFiles)
	fmt.Printf("  Tracked files: %d\n", summary.TrackedFiles)
	fmt.Printf("  Untracked files: %d\n", summary.UntrackedFiles)
	fmt.Printf("  Modified files: %d\n", summary.ModifiedFiles)
	fmt.Printf("  Staged files: %d\n", summary.StagedFiles)
	fmt.Println()

	if summary.ModifiedFiles > 0 || summary.UntrackedFiles > 0 {
		fmt.Println("  Changed files:")
		for _, status := range summary.FileStatuses {
			if status.StatusCode != "" {
				fmt.Printf("    %s %s\n", status.StatusCode, status.Path)
			}
		}
	}

	return nil
}

// Helper functions

// isTracked checks if a file is tracked by Git
func isTracked(trackedFiles []string, path string) bool {
	for _, trackedPath := range trackedFiles {
		if trackedPath == path {
			return true
		}
	}
	return false
}

// isModified checks if a file is modified
func isModified(statusCode string) bool {
	return strings.Contains(statusCode, "M") || strings.Contains(statusCode, "D") || strings.Contains(statusCode, "R")
}

// isStaged checks if a file is staged
func isStaged(statusCode string) bool {
	if len(statusCode) < 1 {
		return false
	}
	return statusCode[0] != ' ' && statusCode[0] != '?'
}

// getStatusDescription returns a human-readable description of a status code
func getStatusDescription(statusCode string) string {
	if statusCode == "??" {
		return "Untracked"
	}
	if statusCode == "A " {
		return "Added"
	}
	if statusCode == " M" {
		return "Modified (not staged)"
	}
	if statusCode == "M " {
		return "Modified (staged)"
	}
	if statusCode == "MM" {
		return "Modified (partially staged)"
	}
	if statusCode == "D " {
		return "Deleted (staged)"
	}
	if statusCode == " D" {
		return "Deleted (not staged)"
	}
	if statusCode == "R " {
		return "Renamed"
	}
	if statusCode == "C " {
		return "Copied"
	}
	if statusCode == "UU" {
		return "Conflict"
	}
	return "Unknown status"
}

// getRepositoryState returns a human-readable description of the repository state
func getRepositoryState(isDirty bool) string {
	if isDirty {
		return "Dirty (uncommitted changes)"
	}
	return "Clean (no uncommitted changes)"
}
