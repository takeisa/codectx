package formatter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"codectx/internal/git"
)

// JSONOutput represents the structure of the JSON output
type JSONOutput struct {
	Metadata      JSONMetadata   `json:"metadata"`
	DirectoryTree string         `json:"directory_tree"`
	Files         []JSONFileInfo `json:"files"`
}

// JSONMetadata contains metadata about the scan
type JSONMetadata struct {
	TargetDirectory  string          `json:"target_directory"`
	ScanTime         string          `json:"scan_time"`
	TotalFiles       int             `json:"total_files"`
	TotalDirectories int             `json:"total_directories"`
	TotalSizeBytes   int64           `json:"total_size_bytes"`
	EstimatedTokens  int             `json:"estimated_tokens"`
	TextFiles        int             `json:"text_files"`
	BinaryFiles      int             `json:"binary_files"`
	ProcessingTime   string          `json:"processing_time,omitempty"`
	Options          JSONScanOptions `json:"options"`
	GitInfo          *git.GitInfo    `json:"git_info,omitempty"`
	Truncated        bool            `json:"truncated,omitempty"`
}

// JSONScanOptions contains information about the scan options
type JSONScanOptions struct {
	IncludeLineNumbers bool     `json:"include_line_numbers"`
	ExtensionsFilter   []string `json:"extensions_filter,omitempty"`
	ExcludePatterns    []string `json:"exclude_patterns,omitempty"`
	Format             string   `json:"format"`
	MaxFileSize        string   `json:"max_file_size,omitempty"`
	CharacterLimit     int64    `json:"character_limit,omitempty"`
	IncludeDotfiles    bool     `json:"include_dotfiles"`
}

// JSONFileInfo contains information about a file
type JSONFileInfo struct {
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
	Type         string `json:"type"`
	SizeBytes    int64  `json:"size_bytes"`
	LineCount    int    `json:"line_count"`
	Extension    string `json:"extension"`
	Content      string `json:"content"`
	Skipped      bool   `json:"skipped,omitempty"`
	SkipReason   string `json:"skip_reason,omitempty"`
	Truncated    bool   `json:"truncated,omitempty"`
}

// formatTreeJSON formats the directory tree in JSON format
func (f *Formatter) formatTreeJSON(tree string) error {
	// Store the tree for later use when we output the full JSON
	metadata := JSONMetadata{
		ScanTime: time.Now().Format(time.RFC3339),
		Options: JSONScanOptions{
			IncludeLineNumbers: f.ShowLineNumbers,
		},
	}

	// Add Git information if available
	if f.GitInfo != nil {
		metadata.GitInfo = f.GitInfo
	}

	f.jsonOutput = &JSONOutput{
		Metadata:      metadata,
		DirectoryTree: tree,
		Files:         []JSONFileInfo{},
	}
	return nil
}

// formatFileContentJSON formats the content of a file in JSON format
func (f *Formatter) formatFileContentJSON(path, relativePath string) error {
	// Get file info
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Read file content
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Count lines
	lineCount := 0
	for _, b := range content {
		if b == '\n' {
			lineCount++
		}
	}
	// Add one for the last line if it doesn't end with a newline
	if len(content) > 0 && content[len(content)-1] != '\n' {
		lineCount++
	}

	// Get file extension
	ext := filepath.Ext(path)
	if ext != "" {
		// Remove the leading dot
		ext = ext[1:]
	}

	// Add file info to the JSON output
	fileEntry := JSONFileInfo{
		Path:         path,
		RelativePath: relativePath,
		Type:         "text",
		SizeBytes:    fileInfo.Size(),
		LineCount:    lineCount,
		Extension:    ext,
		Content:      string(content),
	}

	if f.jsonOutput != nil {
		f.jsonOutput.Files = append(f.jsonOutput.Files, fileEntry)
		f.jsonOutput.Metadata.TotalFiles++
		f.jsonOutput.Metadata.TotalSizeBytes += fileEntry.SizeBytes
		f.jsonOutput.Metadata.EstimatedTokens += len(content) / 4 // Rough estimate
	}

	return nil
}

// finalizeJSON writes the complete JSON output
func (f *Formatter) finalizeJSON() error {
	if f.jsonOutput == nil {
		return fmt.Errorf("no JSON output to finalize")
	}

	// Marshal the JSON output
	jsonData, err := json.MarshalIndent(f.jsonOutput, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write the JSON output
	_, err = fmt.Fprintln(f.Writer, string(jsonData))
	return err
}
