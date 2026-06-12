package evaluator

import (
	"path/filepath"
	"strings"
	"unicode"
)

// nonCodeExtensions are file extensions that are not source code.
var nonCodeExtensions = map[string]bool{
	".md":    true,
	".txt":   true,
	".json":  true,
	".yml":   true,
	".yaml":  true,
	".toml":  true,
	".xml":   true,
	".html":  true,
	".css":   true,
	".scss":  true,
	".lock":  true,
	".sum":   true,
	".svg":   true,
	".png":   true,
	".jpg":   true,
	".gif":   true,
	".ico":   true,
	".woff":  true,
	".woff2": true,
	".ttf":   true,
	".eot":   true,
}

// nonCodeFilenames are files that are not source code regardless of extension.
var nonCodeFilenames = map[string]bool{
	"LICENSE":       true,
	"LICENSE.md":    true,
	"LICENSE.txt":   true,
	".gitignore":    true,
	".editorconfig": true,
	"Makefile":      true,
	"Dockerfile":    true,
	"docker-compose.yml": true,
	"docker-compose.yaml": true,
}

// IsTrivialChange returns true if the file or diff represents a trivial change
// that does not require quality analysis.
// - Non-code files are always trivial
// - Code files with whitespace-only diff are trivial
// - Code files with no diff (empty string) are not trivial (file still needs analysis)
func IsTrivialChange(filePath string, diff string) bool {
	// Check if it's a non-code file
	if isNonCodeFile(filePath) {
		return true
	}

	// Code file with no diff: not trivial (file still exists and needs analysis)
	if diff == "" {
		return false
	}

	// Check if the diff is whitespace-only
	if isWhitespaceOnly(diff) {
		return true
	}

	return false
}

// isNonCodeFile returns true if the file is not a source code file.
func isNonCodeFile(path string) bool {
	base := filepath.Base(path)

	// Check known non-code filenames
	if nonCodeFilenames[base] {
		return true
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(path))
	if nonCodeExtensions[ext] {
		return true
	}

	return false
}

// isWhitespaceOnly returns true if the string contains only whitespace.
func isWhitespaceOnly(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return len(s) > 0
}

// TrivialChangeMessage returns a human-readable message explaining the skip.
func TrivialChangeMessage(filePath string) string {
	return "Changes are trivial (non-code or whitespace-only): " + filePath
}
