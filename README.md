# codectx

A unified directory and file content viewer CLI tool.

*Read this in [Japanese](./README_ja.md)*

## Overview

Codectx is a command-line tool that recursively scans files in a local directory and outputs the directory structure and file contents in an integrated text format. It serves as a CLI version of https://uithub.com/, providing local functionality for formatting repositories for AI analysis.

## Features

- **Directory Scanning**: Recursively explore files and directories from a specified location
- **Tree Visualization**: Display directory structure in a tree format
- **File Content Output**: Show file contents with line numbers
- **Multiple Output Formats**: Support for text (default), HTML, Markdown, and JSON
- **Filtering Capabilities**: Filter by file extensions, exclude patterns, and more
- **Git Integration**: Respect .gitignore, show Git status, and work with Git-tracked files
- **Advanced Analysis**: Perform health checks, complexity analysis, and language statistics
- **Size Limitations**: Control output size with character limits and file size restrictions

## Installation

```bash
# Clone the repository
git clone https://github.com/takeisa/codectx.git

# Build the project
cd codectx
go build
```

## Usage

```bash
codectx [TARGET_DIR] [OPTIONS]
```

### Basic Examples

```bash
codectx foo           # Scan the foo directory
codectx foo/bar       # Scan the foo/bar directory
codectx               # Scan the current directory
codectx .             # Explicitly scan the current directory
```

### Options

#### Output Format
```bash
-f, --format <FORMAT>    Specify output format (text, html, markdown, json)
```

#### File Filtering
```bash
-e, --extensions <EXT1,EXT2,...>    Filter by file extensions (comma-separated)
-x, --exclude <PATTERN1,PATTERN2,...>    Exclude patterns (comma-separated)
--include-dotfiles                  Include dotfiles (default: excluded)
```

#### Size Limits
```bash
-l, --limit <NUMBER>    Maximum character limit (0 for no limit)
--max-file-size <SIZE>  Maximum file size (default: 1MB)
```

#### Other Options
```bash
-o, --output <FILE>     Specify output file (default: stdout)
-n, --no-line-numbers   Don't show line numbers
-v, --verbose           Verbose output mode
-h, --help              Show help
--version               Show version
--dry-run               Show files without processing
```

#### Git Integration
```bash
--git-only              Only include Git tracked files
--respect-gitignore     Respect .gitignore patterns
--ignore-gitignore      Ignore .gitignore patterns (default)
--include-git-info      Include Git information in output
--git-status            Show Git status information
```

#### Advanced Analysis
```bash
--health-check          Perform project health check
--complexity-analysis   Perform complexity analysis
--language-stats        Show language statistics
```

## Use Cases

### AI Code Explanation
```bash
# Explain TypeScript project to AI
codectx -e ts,tsx,json -l 100000 > project_context.txt
```

### Code Review Preparation
```bash
# Output only changed files in Markdown
codectx --git-only --format markdown -o review.md
```

### Documentation Generation
```bash
# Output project structure in Markdown
codectx --format markdown -o project_structure.md
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
