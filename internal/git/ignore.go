package git

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// GitIgnoreParser parses .gitignore files and checks if files should be ignored
type GitIgnoreParser struct {
	patterns []string
	rules    []GitIgnoreRule
	rootDir  string
}

// GitIgnoreRule represents a single rule in a .gitignore file
type GitIgnoreRule struct {
	Pattern     string
	IsNegation  bool // ! で始まる場合
	IsDirectory bool // / で終わる場合
}

// NewGitIgnoreParser creates a new GitIgnoreParser
func NewGitIgnoreParser(rootDir string) *GitIgnoreParser {
	return &GitIgnoreParser{
		rootDir: rootDir,
	}
}

// ParseGitIgnore parses a .gitignore file and adds its rules to the parser
func (g *GitIgnoreParser) ParseGitIgnore(gitignorePath string) error {
	file, err := os.Open(gitignorePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		g.patterns = append(g.patterns, line)

		rule := GitIgnoreRule{
			Pattern: line,
		}

		// Check if it's a negation pattern
		if strings.HasPrefix(line, "!") {
			rule.IsNegation = true
			rule.Pattern = line[1:]
		}

		// Check if it's a directory pattern
		if strings.HasSuffix(rule.Pattern, "/") {
			rule.IsDirectory = true
			rule.Pattern = rule.Pattern[:len(rule.Pattern)-1]
		}

		g.rules = append(g.rules, rule)
	}

	return scanner.Err()
}

// ParseAllGitIgnores finds and parses all .gitignore files in the repository
func (g *GitIgnoreParser) ParseAllGitIgnores() error {
	return filepath.Walk(g.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Base(path) == ".gitignore" {
			if err := g.ParseGitIgnore(path); err != nil {
				return err
			}
		}

		return nil
	})
}

// ShouldIgnore checks if a file should be ignored based on .gitignore rules
func (g *GitIgnoreParser) ShouldIgnore(filePath string) bool {
	// Make the path relative to the root directory
	relPath, err := filepath.Rel(g.rootDir, filePath)
	if err != nil {
		return false
	}

	// Normalize path separators
	relPath = filepath.ToSlash(relPath)

	// Check each rule in reverse order (later rules override earlier ones)
	for i := len(g.rules) - 1; i >= 0; i-- {
		rule := g.rules[i]

		// Check if the pattern matches
		matched, _ := filepath.Match(rule.Pattern, relPath)
		if !matched {
			// Also check if the pattern matches any part of the path
			parts := strings.Split(relPath, "/")
			for j := 0; j < len(parts); j++ {
				subPath := strings.Join(parts[j:], "/")
				matched, _ = filepath.Match(rule.Pattern, subPath)
				if matched {
					break
				}
			}
		}

		if matched {
			// If it's a negation rule, don't ignore
			if rule.IsNegation {
				return false
			}

			// If it's a directory rule, only ignore if the path is a directory
			if rule.IsDirectory {
				info, err := os.Stat(filePath)
				if err != nil || !info.IsDir() {
					continue
				}
			}

			return true
		}
	}

	return false
}

// IsGitAvailable checks if git is available on the system
func IsGitAvailable() bool {
	_, err := os.Stat(filepath.Join(".git"))
	return err == nil
}
