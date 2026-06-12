package scanner

import (
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

// Scanner walks a directory tree and returns files that are not excluded
// by the configured glob patterns.
type Scanner struct {
	exclusions []string
}

// New creates a Scanner with the given exclusion glob patterns.
func New(exclusions []string) *Scanner {
	return &Scanner{
		exclusions: exclusions,
	}
}

// Scan walks the root directory and returns all non-excluded file paths
// and any warnings for exclusion patterns that matched zero files.
func (s *Scanner) Scan(root string) ([]string, []string) {
	var files []string
	matchedExclusions := make(map[string]bool)

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		if s.isExcluded(rel) {
			// Mark which exclusion patterns matched
			for _, pattern := range s.exclusions {
				if m, _ := doublestar.Match(pattern, rel); m {
					matchedExclusions[pattern] = true
				}
			}
			return nil
		}

		files = append(files, path)
		return nil
	})

	// Generate warnings for patterns that matched nothing
	var warnings []string
	for _, pattern := range s.exclusions {
		if !matchedExclusions[pattern] {
			warnings = append(warnings, "exclusion pattern matched no files: "+pattern)
		}
	}

	return files, warnings
}

// isExcluded checks if a relative path matches any exclusion pattern.
func (s *Scanner) isExcluded(rel string) bool {
	for _, pattern := range s.exclusions {
		if matched, err := doublestar.Match(pattern, rel); matched && err == nil {
			return true
		}
	}
	return false
}
