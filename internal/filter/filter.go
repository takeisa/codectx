package filter

import (
	"path/filepath"
	"strings"

	"codectx/internal/git"
)

// Filter defines criteria for including or excluding files
type Filter struct {
	Extensions      []string
	ExcludePatterns []string
	IncludeDotfiles bool
	GitIgnoreParser *git.GitIgnoreParser
	GitTrackedOnly  bool
	GitTrackedFiles []string
}

// NewFilter creates a new filter with the given criteria
func NewFilter(extensions, excludePatterns string, includeDotfiles bool) *Filter {
	var exts []string
	if extensions != "" {
		exts = strings.Split(extensions, ",")
		// Normalize extensions to have a leading dot
		for i, ext := range exts {
			ext = strings.TrimSpace(ext)
			if !strings.HasPrefix(ext, ".") {
				exts[i] = "." + ext
			} else {
				exts[i] = ext
			}
		}
	}

	var patterns []string
	if excludePatterns != "" {
		patterns = strings.Split(excludePatterns, ",")
		for i, pattern := range patterns {
			patterns[i] = strings.TrimSpace(pattern)
		}
	}

	return &Filter{
		Extensions:      exts,
		ExcludePatterns: patterns,
		IncludeDotfiles: includeDotfiles,
	}
}

// SetGitIgnoreParser sets the GitIgnoreParser for the filter
func (f *Filter) SetGitIgnoreParser(parser *git.GitIgnoreParser) {
	f.GitIgnoreParser = parser
}

// SetGitTrackedFiles sets the list of Git tracked files and enables Git tracked only mode
func (f *Filter) SetGitTrackedFiles(files []string) {
	f.GitTrackedFiles = files
	f.GitTrackedOnly = true
}

// ShouldInclude determines if a file should be included based on the filter criteria
func (f *Filter) ShouldInclude(path string) bool {
	// Get the base name of the file
	base := filepath.Base(path)

	// Check if it's a dotfile
	if !f.IncludeDotfiles && strings.HasPrefix(base, ".") {
		return false
	}

	// Check if we should only include Git tracked files
	if f.GitTrackedOnly && len(f.GitTrackedFiles) > 0 {
		isTracked := false
		relPath := path
		for _, trackedPath := range f.GitTrackedFiles {
			if trackedPath == relPath || filepath.Join(filepath.Dir(path), trackedPath) == path {
				isTracked = true
				break
			}
		}
		if !isTracked {
			return false
		}
	}

	// Check if the file should be ignored based on .gitignore rules
	if f.GitIgnoreParser != nil && f.GitIgnoreParser.ShouldIgnore(path) {
		return false
	}

	// Check exclusion patterns
	for _, pattern := range f.ExcludePatterns {
		matched, err := filepath.Match(pattern, base)
		if err == nil && matched {
			return false
		}

		// Also check if the pattern matches the full path
		matched, err = filepath.Match(pattern, path)
		if err == nil && matched {
			return false
		}
	}

	// If no extensions are specified, include all files
	if len(f.Extensions) == 0 {
		return true
	}

	// Check if the file has one of the specified extensions
	ext := filepath.Ext(path)
	for _, allowedExt := range f.Extensions {
		if ext == allowedExt {
			return true
		}
	}

	// If we have extension filters and none matched, exclude the file
	return false
}
