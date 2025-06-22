package formatter

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

// formatFileContentMarkdown formats the content of a file in Markdown format
func (f *Formatter) formatFileContentMarkdown(path, relativePath string) error {
	// Print the file header
	fmt.Fprintf(f.Writer, "\n### %s\n", relativePath)

	// If the file has a specific extension, add it to the code block with proper language identifier
	ext := filepath.Ext(relativePath)
	langId := getLanguageIdentifier(ext)
	fmt.Fprintf(f.Writer, "```%s\n", langId)

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		line := scanner.Text()
		if f.ShowLineNumbers {
			fmt.Fprintf(f.Writer, "%d | %s\n", lineNum, line)
		} else {
			fmt.Fprintln(f.Writer, line)
		}
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Close the code block
	fmt.Fprintln(f.Writer, "```")

	return nil
}

// formatTreeMarkdown formats the directory tree in Markdown format
func (f *Formatter) formatTreeMarkdown(tree string) error {
	fmt.Fprintln(f.Writer, "# Project Structure")
	fmt.Fprintln(f.Writer, "")
	fmt.Fprintln(f.Writer, "## Directory Tree")
	fmt.Fprintln(f.Writer, "```")
	fmt.Fprintln(f.Writer, tree)
	fmt.Fprintln(f.Writer, "```")
	fmt.Fprintln(f.Writer, "")
	fmt.Fprintln(f.Writer, "## Files")
	return nil
}

// getLanguageIdentifier returns the appropriate language identifier for syntax highlighting
func getLanguageIdentifier(ext string) string {
	if ext == "" {
		return ""
	}

	// Remove the leading dot
	ext = ext[1:]

	// Map common extensions to language identifiers
	langMap := map[string]string{
		"go":         "go",
		"js":         "javascript",
		"ts":         "typescript",
		"py":         "python",
		"java":       "java",
		"c":          "c",
		"cpp":        "cpp",
		"cc":         "cpp",
		"cxx":        "cpp",
		"h":          "c",
		"hpp":        "cpp",
		"cs":         "csharp",
		"php":        "php",
		"rb":         "ruby",
		"rs":         "rust",
		"kt":         "kotlin",
		"swift":      "swift",
		"scala":      "scala",
		"sh":         "bash",
		"bash":       "bash",
		"zsh":        "bash",
		"fish":       "bash",
		"ps1":        "powershell",
		"sql":        "sql",
		"html":       "html",
		"htm":        "html",
		"xml":        "xml",
		"css":        "css",
		"scss":       "scss",
		"sass":       "sass",
		"less":       "less",
		"json":       "json",
		"yaml":       "yaml",
		"yml":        "yaml",
		"toml":       "toml",
		"ini":        "ini",
		"cfg":        "ini",
		"conf":       "ini",
		"md":         "markdown",
		"txt":        "text",
		"log":        "text",
		"gitignore":  "gitignore",
		"dockerfile": "dockerfile",
		"makefile":   "makefile",
	}

	if lang, exists := langMap[ext]; exists {
		return lang
	}

	return ext
}
