package formatter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"codectx/internal/git"
	"codectx/internal/limits"
)

// OutputFormat represents the format of the output
type OutputFormat string

const (
	// TextFormat is the default plain text format
	TextFormat OutputFormat = "text"
	// HTMLFormat is HTML format
	HTMLFormat OutputFormat = "html"
	// MarkdownFormat is Markdown format
	MarkdownFormat OutputFormat = "markdown"
	// JSONFormat is JSON format
	JSONFormat OutputFormat = "json"
)

// Formatter handles the formatting of the output
type Formatter struct {
	Format          OutputFormat
	ShowLineNumbers bool
	Writer          io.Writer
	jsonOutput      *JSONOutput
	SizeLimiter     *limits.SizeLimiter
	GitInfo         *git.GitInfo
}

// NewFormatter creates a new formatter with the given format
func NewFormatter(format string, showLineNumbers bool, outputPath string, sizeLimiter *limits.SizeLimiter, gitInfo *git.GitInfo) (*Formatter, error) {
	var outputFormat OutputFormat
	switch strings.ToLower(format) {
	case "text":
		outputFormat = TextFormat
	case "html":
		outputFormat = HTMLFormat
	case "markdown":
		outputFormat = MarkdownFormat
	case "json":
		outputFormat = JSONFormat
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	var writer io.Writer = os.Stdout
	if outputPath != "" {
		file, err := os.Create(outputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create output file: %w", err)
		}
		writer = file
	}

	return &Formatter{
		Format:          outputFormat,
		ShowLineNumbers: showLineNumbers,
		Writer:          writer,
		SizeLimiter:     sizeLimiter,
		GitInfo:         gitInfo,
	}, nil
}

// FormatTree formats the directory tree
func (f *Formatter) FormatTree(tree string) error {
	switch f.Format {
	case TextFormat:
		_, err := fmt.Fprintln(f.Writer, tree)
		return err
	case MarkdownFormat:
		return f.formatTreeMarkdown(tree)
	case JSONFormat:
		return f.formatTreeJSON(tree)
	case HTMLFormat:
		return f.formatTreeHTML(tree)
	default:
		return fmt.Errorf("format not implemented: %s", f.Format)
	}
}

// FormatFileContent formats the content of a file
func (f *Formatter) FormatFileContent(path, relativePath string) error {
	switch f.Format {
	case TextFormat:
		return f.formatFileContentText(path, relativePath)
	case MarkdownFormat:
		return f.formatFileContentMarkdown(path, relativePath)
	case JSONFormat:
		return f.formatFileContentJSON(path, relativePath)
	case HTMLFormat:
		return f.formatFileContentHTML(path, relativePath)
	default:
		return fmt.Errorf("format not implemented: %s", f.Format)
	}
}

// formatFileContentText formats the content of a file in text format
func (f *Formatter) formatFileContentText(path, relativePath string) error {
	// Check if we have a size limiter
	if f.SizeLimiter != nil {
		// Check if the file exceeds the maximum file size
		withinLimit, fileSize, err := f.SizeLimiter.CheckFileSize(path)
		if err != nil {
			return fmt.Errorf("failed to check file size: %w", err)
		}

		if !withinLimit {
			// File is too large, print a message instead of the content
			fmt.Fprintf(f.Writer, "\n%s:\n", relativePath)
			fmt.Fprintln(f.Writer, "--------------------------------------------------------------------------------")
			fmt.Fprintln(f.Writer, f.SizeLimiter.GetFileTooLargeMessage(path, fileSize))
			return nil
		}
	}

	// Print the file header
	fmt.Fprintf(f.Writer, "\n%s:\n", relativePath)
	fmt.Fprintln(f.Writer, "--------------------------------------------------------------------------------")

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

		// Format the line
		var formattedLine string
		if f.ShowLineNumbers {
			formattedLine = fmt.Sprintf("%2d | %s\n", lineNum, line)
		} else {
			formattedLine = line + "\n"
		}

		// Check if adding this line would exceed the total size limit
		if f.SizeLimiter != nil && f.SizeLimiter.MaxTotalSize > 0 {
			if !f.SizeLimiter.AddToTotalSize(int64(len(formattedLine))) {
				// We've reached the limit, print a message and stop
				fmt.Fprintln(f.Writer, f.SizeLimiter.GetTruncatedMessage())
				return nil
			}
		}

		// Write the line
		fmt.Fprint(f.Writer, formattedLine)
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}

// Finalize performs any final operations needed for the formatter
func (f *Formatter) Finalize() error {
	switch f.Format {
	case HTMLFormat:
		return f.finalizeHTML()
	case JSONFormat:
		return f.finalizeJSON()
	}
	return nil
}

// Close closes any resources used by the formatter
func (f *Formatter) Close() error {
	// First finalize the output if needed
	if err := f.Finalize(); err != nil {
		return err
	}

	// Then close the writer if it's closable
	if closer, ok := f.Writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
