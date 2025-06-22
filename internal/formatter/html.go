package formatter

import (
	"bufio"
	"fmt"
	"html"
	"os"
	"strings"
)

// HTML template constants
const (
	htmlHeader = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Project Structure</title>
    <style>
        body { 
            font-family: 'Courier New', monospace; 
            line-height: 1.6; 
            margin: 0; 
            padding: 20px; 
            background-color: #f8f9fa;
        }
        .container { 
            max-width: 1200px; 
            margin: 0 auto; 
            background: white; 
            padding: 20px; 
            border-radius: 8px; 
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 { 
            color: #333; 
            border-bottom: 2px solid #007acc; 
            padding-bottom: 10px; 
        }
        .tree { 
            background: #f5f5f5; 
            padding: 15px; 
            border-radius: 4px; 
            border-left: 4px solid #007acc; 
            margin: 20px 0; 
            font-size: 14px;
        }
        .file { 
            margin: 20px 0; 
            border: 1px solid #ddd; 
            border-radius: 4px; 
            overflow: hidden;
        }
        .file-header { 
            background: #e9ecef; 
            padding: 10px 15px; 
            font-weight: bold; 
            color: #495057; 
            border-bottom: 1px solid #ddd;
        }
        .file-content { 
            background: #f8f9fa; 
            padding: 15px; 
            white-space: pre-wrap; 
            font-family: 'Courier New', monospace; 
            font-size: 13px; 
            line-height: 1.5; 
            overflow-x: auto;
        }
        .line-number { 
            color: #6c757d; 
            margin-right: 15px; 
            user-select: none; 
            display: inline-block; 
            width: 30px; 
            text-align: right;
        }
        .line { 
            display: block; 
        }
        .metadata { 
            background: #e3f2fd; 
            padding: 10px; 
            border-radius: 4px; 
            margin: 20px 0; 
            font-size: 12px; 
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Project Structure</h1>
        <div class="tree">%s</div>
        <div class="files">
`

	htmlFooter = `        </div>
    </div>
</body>
</html>
`

	htmlFileHeader = `        <div class="file">
            <div class="file-header">%s</div>
            <div class="file-content">
`

	htmlFileFooter = `            </div>
        </div>
`
)

// formatTreeHTML formats the directory tree in HTML format
func (f *Formatter) formatTreeHTML(tree string) error {
	// Escape the tree for HTML
	escapedTree := html.EscapeString(tree)
	// Replace newlines with <br> tags
	escapedTree = strings.ReplaceAll(escapedTree, "\n", "<br>")

	// Write the HTML header with the tree
	_, err := fmt.Fprintf(f.Writer, htmlHeader, escapedTree)
	return err
}

// formatFileContentHTML formats the content of a file in HTML format
func (f *Formatter) formatFileContentHTML(path, relativePath string) error {
	// Write the file header
	_, err := fmt.Fprintf(f.Writer, htmlFileHeader, html.EscapeString(relativePath))
	if err != nil {
		return err
	}

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
		// Escape the line for HTML
		escapedLine := html.EscapeString(line)

		if f.ShowLineNumbers {
			_, err = fmt.Fprintf(f.Writer, "<span class=\"line\"><span class=\"line-number\">%d</span>%s</span>\n", lineNum, escapedLine)
		} else {
			_, err = fmt.Fprintf(f.Writer, "<span class=\"line\">%s</span>\n", escapedLine)
		}

		if err != nil {
			return err
		}

		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Write the file footer
	_, err = fmt.Fprint(f.Writer, htmlFileFooter)
	return err
}

// finalizeHTML writes the HTML footer
func (f *Formatter) finalizeHTML() error {
	_, err := fmt.Fprint(f.Writer, htmlFooter)
	return err
}
