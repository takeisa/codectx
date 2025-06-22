package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileEntry represents a file or directory in the scanned structure
type FileEntry struct {
	Path     string
	IsDir    bool
	Children []*FileEntry
}

// Scanner handles directory scanning and tree generation
type Scanner struct {
	RootDir         string
	IncludeDotfiles bool
}

// NewScanner creates a new scanner for the given directory
func NewScanner(rootDir string, includeDotfiles bool) *Scanner {
	return &Scanner{
		RootDir:         rootDir,
		IncludeDotfiles: includeDotfiles,
	}
}

// Scan performs the directory scan and returns the root entry
func (s *Scanner) Scan() (*FileEntry, error) {
	rootInfo, err := os.Stat(s.RootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to access root directory: %w", err)
	}

	if !rootInfo.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", s.RootDir)
	}

	root := &FileEntry{
		Path:  s.RootDir,
		IsDir: true,
	}

	err = s.scanDir(root)
	if err != nil {
		return nil, err
	}

	return root, nil
}

// scanDir recursively scans a directory and populates the children of the given entry
func (s *Scanner) scanDir(entry *FileEntry) error {
	entries, err := os.ReadDir(entry.Path)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", entry.Path, err)
	}

	for _, dirEntry := range entries {
		name := dirEntry.Name()

		// Skip dotfiles if not explicitly included
		if !s.IncludeDotfiles && strings.HasPrefix(name, ".") {
			continue
		}

		path := filepath.Join(entry.Path, name)
		isDir := dirEntry.IsDir()

		child := &FileEntry{
			Path:  path,
			IsDir: isDir,
		}

		if isDir {
			if err := s.scanDir(child); err != nil {
				// Just log the error and continue if we can't access a subdirectory
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
				continue
			}
		}

		entry.Children = append(entry.Children, child)
	}

	// Sort children: directories first, then files, both alphabetically
	sort.Slice(entry.Children, func(i, j int) bool {
		if entry.Children[i].IsDir != entry.Children[j].IsDir {
			return entry.Children[i].IsDir
		}
		return filepath.Base(entry.Children[i].Path) < filepath.Base(entry.Children[j].Path)
	})

	return nil
}

// GenerateTree creates a string representation of the directory tree
func (s *Scanner) GenerateTree(root *FileEntry) string {
	var sb strings.Builder
	s.generateTreeRecursive(&sb, root, "", true)
	return sb.String()
}

// generateTreeRecursive builds the tree representation recursively
func (s *Scanner) generateTreeRecursive(sb *strings.Builder, entry *FileEntry, prefix string, isLast bool) {
	// Skip the root directory itself
	if entry.Path != s.RootDir {
		if isLast {
			sb.WriteString(prefix + "└── ")
			prefix += "    "
		} else {
			sb.WriteString(prefix + "├── ")
			prefix += "│   "
		}

		// Write the entry name
		sb.WriteString(filepath.Base(entry.Path))
		if entry.IsDir {
			sb.WriteString("/")
		}
		sb.WriteString("\n")
	}

	// Process children
	for i, child := range entry.Children {
		isLastChild := i == len(entry.Children)-1
		s.generateTreeRecursive(sb, child, prefix, isLastChild)
	}
}

// GetRelativePaths returns a list of all file paths relative to the root directory
func (s *Scanner) GetRelativePaths(root *FileEntry) []string {
	var paths []string
	s.collectRelativePaths(root, &paths)
	return paths
}

// collectRelativePaths recursively collects relative paths from the given entry
func (s *Scanner) collectRelativePaths(entry *FileEntry, paths *[]string) {
	// Skip directories
	if !entry.IsDir {
		relPath, err := filepath.Rel(s.RootDir, entry.Path)
		if err == nil {
			*paths = append(*paths, "/"+relPath)
		}
	}

	// Process children
	for _, child := range entry.Children {
		s.collectRelativePaths(child, paths)
	}
}
