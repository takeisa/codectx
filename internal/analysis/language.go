package analysis

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LanguageStats represents the language statistics for a project
type LanguageStats struct {
	TotalFiles   int                     `json:"total_files"`
	TotalSize    int64                   `json:"total_size"`
	Languages    map[string]LanguageInfo `json:"languages"`
	TopLanguages []LanguageInfo          `json:"top_languages"`
}

// LanguageInfo contains information about a language
type LanguageInfo struct {
	Name       string   `json:"name"`
	Files      int      `json:"files"`
	Size       int64    `json:"size"`
	Percentage float64  `json:"percentage"`
	Extensions []string `json:"extensions"`
}

// NewLanguageStats creates a new language statistics
func NewLanguageStats() *LanguageStats {
	return &LanguageStats{
		Languages: make(map[string]LanguageInfo),
	}
}

// AnalyzeLanguages performs a language analysis on the project
func AnalyzeLanguages(rootDir string) (*LanguageStats, error) {
	stats := NewLanguageStats()

	// Map file extensions to languages
	extToLang := getExtensionToLanguageMap()

	// Track extensions for each language
	langToExts := make(map[string]map[string]bool)

	// Walk the directory tree to analyze files
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if info.IsDir() && filepath.Base(path) == ".git" {
			return filepath.SkipDir
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get file extension
		ext := strings.ToLower(filepath.Ext(path))
		if ext == "" {
			return nil
		}

		// Remove the leading dot
		ext = ext[1:]

		// Get language for this extension
		lang, ok := extToLang[ext]
		if !ok {
			lang = "Other"
		}

		// Update language info
		if langInfo, ok := stats.Languages[lang]; ok {
			langInfo.Files++
			langInfo.Size += info.Size()
			stats.Languages[lang] = langInfo
		} else {
			stats.Languages[lang] = LanguageInfo{
				Name:  lang,
				Files: 1,
				Size:  info.Size(),
			}
		}

		// Track extensions for this language
		if _, ok := langToExts[lang]; !ok {
			langToExts[lang] = make(map[string]bool)
		}
		langToExts[lang][ext] = true

		// Update total stats
		stats.TotalFiles++
		stats.TotalSize += info.Size()

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to analyze languages: %w", err)
	}

	// Calculate percentages and collect extensions
	for lang, info := range stats.Languages {
		if stats.TotalFiles > 0 {
			info.Percentage = float64(info.Files) / float64(stats.TotalFiles) * 100
		}

		// Collect extensions for this language
		for ext := range langToExts[lang] {
			info.Extensions = append(info.Extensions, ext)
		}

		// Sort extensions
		sort.Strings(info.Extensions)

		stats.Languages[lang] = info
	}

	// Create sorted list of top languages
	for _, info := range stats.Languages {
		stats.TopLanguages = append(stats.TopLanguages, info)
	}

	// Sort by number of files (descending)
	sort.Slice(stats.TopLanguages, func(i, j int) bool {
		return stats.TopLanguages[i].Files > stats.TopLanguages[j].Files
	})

	return stats, nil
}

// PrintLanguageStats prints the language statistics
func PrintLanguageStats(stats *LanguageStats) {
	fmt.Println("\nLanguage Statistics:")
	fmt.Println("====================")

	fmt.Printf("\nTotal files: %d\n", stats.TotalFiles)
	fmt.Printf("Total size: %.2f MB\n", float64(stats.TotalSize)/(1024*1024))

	fmt.Println("\nLanguage Distribution:")
	for _, lang := range stats.TopLanguages {
		fmt.Printf("  %s: %d files (%.1f%%) - %.2f KB\n",
			lang.Name, lang.Files, lang.Percentage, float64(lang.Size)/1024)
	}

	fmt.Println("\nFile Extensions by Language:")
	for _, lang := range stats.TopLanguages {
		if len(lang.Extensions) > 0 {
			fmt.Printf("  %s: %s\n", lang.Name, strings.Join(lang.Extensions, ", "))
		}
	}
}

// getExtensionToLanguageMap returns a map of file extensions to languages
func getExtensionToLanguageMap() map[string]string {
	return map[string]string{
		// Go
		"go": "Go",

		// Web
		"html": "HTML",
		"htm":  "HTML",
		"css":  "CSS",
		"scss": "CSS",
		"sass": "CSS",
		"less": "CSS",
		"js":   "JavaScript",
		"jsx":  "JavaScript",
		"ts":   "TypeScript",
		"tsx":  "TypeScript",

		// C-family
		"c":   "C",
		"h":   "C",
		"cpp": "C++",
		"hpp": "C++",
		"cc":  "C++",
		"cxx": "C++",
		"c++": "C++",
		"cs":  "C#",

		// JVM
		"java":   "Java",
		"kt":     "Kotlin",
		"kts":    "Kotlin",
		"scala":  "Scala",
		"groovy": "Groovy",

		// Python
		"py":  "Python",
		"pyc": "Python",
		"pyd": "Python",
		"pyo": "Python",
		"pyw": "Python",

		// Ruby
		"rb":  "Ruby",
		"erb": "Ruby",

		// PHP
		"php": "PHP",

		// Swift
		"swift": "Swift",

		// Rust
		"rs": "Rust",

		// Shell
		"sh":   "Shell",
		"bash": "Shell",
		"zsh":  "Shell",

		// Data formats
		"json": "JSON",
		"yaml": "YAML",
		"yml":  "YAML",
		"xml":  "XML",
		"toml": "TOML",
		"csv":  "CSV",
		"tsv":  "CSV",

		// Markdown and text
		"md":       "Markdown",
		"markdown": "Markdown",
		"txt":      "Text",
		"rst":      "Text",

		// Config files
		"ini":  "Config",
		"cfg":  "Config",
		"conf": "Config",

		// Documentation
		"pdf":  "Document",
		"doc":  "Document",
		"docx": "Document",
		"odt":  "Document",

		// Images
		"png":  "Image",
		"jpg":  "Image",
		"jpeg": "Image",
		"gif":  "Image",
		"svg":  "Image",
		"webp": "Image",
		"ico":  "Image",

		// Audio
		"mp3": "Audio",
		"wav": "Audio",
		"ogg": "Audio",

		// Video
		"mp4":  "Video",
		"webm": "Video",
		"avi":  "Video",

		// Archives
		"zip": "Archive",
		"tar": "Archive",
		"gz":  "Archive",
		"rar": "Archive",
		"7z":  "Archive",
	}
}
