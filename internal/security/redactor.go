package security

import (
	"regexp"
	"strings"
)

// Redactor detects and redacts sensitive data from diffs.
type Redactor struct {
	patterns []*regexp.Regexp
}

// NewRedactor creates a new redactor with default sensitive data patterns.
func NewRedactor() *Redactor {
	// Common sensitive data patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*["\']?[a-zA-Z0-9_\-]{16,}["\']?`),
		regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*["\']?[^\s"'\n]{4,}["\']?`),
		regexp.MustCompile(`(?i)(secret|secret[_-]?key)\s*[:=]\s*["\']?[a-zA-Z0-9_\-]{8,}["\']?`),
		regexp.MustCompile(`(?i)(access[_-]?token|auth[_-]?token)\s*[:=]\s*["\']?[a-zA-Z0-9_\-]{16,}["\']?`),
		regexp.MustCompile(`(?i)(private[_-]?key)\s*[:=]\s*["\']?[a-zA-Z0-9_\-]{16,}["\']?`),
		regexp.MustCompile(`(?i)(aws[_-]?secret[_-]?access[_-]?key)\s*[:=]\s*["\']?[a-zA-Z0-9/+=]{20,}["\']?`),
		regexp.MustCompile(`(?i)(db[_-]?password|database[_-]?password)\s*[:=]\s*["\']?[^\s"'\n]{4,}["\']?`),
		regexp.MustCompile(`(?i)(connection[_-]?string|conn[_-]?str)\s*[:=]\s*["\']?[^\s"'\n]{10,}["\']?`),
	}

	return &Redactor{patterns: patterns}
}

// Redact scans text for sensitive data and replaces it with [REDACTED].
func (r *Redactor) Redact(text string) (string, int) {
	redactionCount := 0

	for _, pattern := range r.patterns {
		matches := pattern.FindAllStringIndex(text, -1)
		for _, match := range matches {
			// Replace the matched text with [REDACTED]
			text = text[:match[0]] + "[REDACTED]" + text[match[1]:]
			redactionCount++
		}
	}

	return text, redactionCount
}

// RedactDiff redacts sensitive data from a diff string.
func (r *Redactor) RedactDiff(diffText string) (string, int) {
	return r.Redact(diffText)
}

// HasSensitiveData checks if text contains any sensitive data patterns.
func (r *Redactor) HasSensitiveData(text string) bool {
	for _, pattern := range r.patterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

// RedactFile redacts sensitive data from a file's content.
func (r *Redactor) RedactFile(content string) (string, int) {
	return r.Redact(content)
}

// AddPattern adds a custom sensitive data pattern.
func (r *Redactor) AddPattern(pattern string) error {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	r.patterns = append(r.patterns, compiled)
	return nil
}

// RedactionReport provides details about redactions performed.
type RedactionReport struct {
	TotalRedactions int               `json:"total_redactions"`
	FilesRedacted   map[string]int    `json:"files_redacted"`
	RedactedLines   map[string][]int  `json:"redacted_lines"`
}

// GenerateReport creates a redaction report for a set of files.
func (r *Redactor) GenerateReport(files map[string]string) RedactionReport {
	report := RedactionReport{
		FilesRedacted: make(map[string]int),
		RedactedLines: make(map[string][]int),
	}

	for filePath, content := range files {
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if r.HasSensitiveData(line) {
				report.TotalRedactions++
				report.FilesRedacted[filePath]++
				report.RedactedLines[filePath] = append(report.RedactedLines[filePath], i+1)
			}
		}
	}

	return report
}
