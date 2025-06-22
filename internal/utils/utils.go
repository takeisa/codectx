package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// IsTextFile checks if a file is a text file by looking at the first 512 bytes
func IsTextFile(path string) (bool, error) {
	// First check by file extension - common text file extensions
	ext := strings.ToLower(filepath.Ext(path))
	textExtensions := []string{
		".txt", ".md", ".markdown", ".rst", ".asciidoc",
		".json", ".xml", ".yaml", ".yml", ".toml", ".ini", ".cfg", ".conf",
		".log", ".csv", ".tsv", ".sql",
		".html", ".htm", ".css", ".js", ".ts", ".jsx", ".tsx",
		".go", ".py", ".java", ".c", ".cpp", ".h", ".hpp", ".cs", ".php", ".rb", ".rs", ".kt", ".swift",
		".sh", ".bash", ".zsh", ".fish", ".ps1", ".bat", ".cmd",
		".dockerfile", ".gitignore", ".gitattributes",
	}
	
	for _, textExt := range textExtensions {
		if ext == textExt {
			return true, nil
		}
	}
	
	// If no extension or unknown extension, check content
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info for size check
	fileInfo, err := file.Stat()
	if err != nil {
		return false, fmt.Errorf("failed to get file info: %w", err)
	}

	// Empty files are considered text
	if fileInfo.Size() == 0 {
		return true, nil
	}

	// Read the first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	// Check for null bytes, which indicate a binary file
	if bytes.Contains(buf[:n], []byte{0}) {
		return false, nil
	}

	// Check if the content is valid UTF-8 (allow partial sequences at the end)
	// For text files, most of the content should be valid UTF-8
	validUTF8Bytes := 0
	for i := 0; i < n; {
		r, size := utf8.DecodeRune(buf[i:n])
		if r == utf8.RuneError && size == 1 {
			// Invalid UTF-8 sequence
			i++
		} else {
			validUTF8Bytes += size
			i += size
		}
	}
	
	// If less than 80% of bytes form valid UTF-8 sequences, consider it binary
	if float64(validUTF8Bytes)/float64(n) < 0.8 {
		return false, nil
	}

	// Check for high concentration of control characters (excluding common whitespace)
	controlChars := 0
	for i := 0; i < n; i++ {
		if buf[i] < 32 && !isPrintableASCII(buf[i]) {
			controlChars++
		}
	}

	// If more than 30% of the first 512 bytes are control characters, consider it binary
	if float64(controlChars)/float64(n) > 0.3 {
		return false, nil
	}

	return true, nil
}

// isPrintableASCII checks if a byte is a printable ASCII character or a common control character
func isPrintableASCII(b byte) bool {
	// Common control characters (newline, tab, etc.)
	return b == '\n' || b == '\r' || b == '\t' || b == '\f' || b == '\v' || b == '\b'
}

// IsBinaryFile checks if a file is a binary file
func IsBinaryFile(path string) (bool, error) {
	isText, err := IsTextFile(path)
	if err != nil {
		return false, err
	}
	return !isText, nil
}
