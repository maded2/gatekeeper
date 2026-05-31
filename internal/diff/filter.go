package diff

import (
	"path/filepath"
	"strings"
)

// Filter removes files matching exclude patterns from a diff.
func Filter(d Diff, patterns []string) (Diff, []string) {
	var filtered []FileChange
	var excluded []string

	for _, f := range d.Files {
		path := f.NewPath
		if path == "" {
			path = f.OldPath
		}
		if matchesPattern(path, patterns) {
			excluded = append(excluded, path)
		} else {
			filtered = append(filtered, f)
		}
	}

	return Diff{
		Files:       filtered,
		TotalLines:  countLines(filtered),
		Description: d.Description,
	}, excluded
}

func matchesPattern(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchSinglePattern(path, pattern) {
			return true
		}
	}
	return false
}

func matchSinglePattern(path, pattern string) bool {
	// Handle glob patterns
	if strings.Contains(pattern, "*") {
		// Use filepath.Match for glob patterns
		if matched, err := filepath.Match(pattern, path); err == nil && matched {
			return true
		}
		// Also check basename for extension patterns
		base := filepath.Base(path)
		if matched, err := filepath.Match(pattern, base); err == nil && matched {
			return true
		}
		// Directory prefix match: vendor/*
		if strings.HasSuffix(pattern, "/*") {
			dir := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(path, dir+"/") {
				return true
			}
		}
	}
	// Exact match
	return path == pattern
}

func countLines(files []FileChange) int {
	total := 0
	for _, f := range files {
		total += strings.Count(f.OldContent, "\n")
		total += strings.Count(f.NewContent, "\n")
	}
	return total
}
