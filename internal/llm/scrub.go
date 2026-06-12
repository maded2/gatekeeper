package llm

import "regexp"

// secretPatterns matches common hardcoded secret patterns.
var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:?]?=\s*["'][^"']{8,}["']`),
	regexp.MustCompile(`(?i)(secret|password|passwd|pwd)\s*[:?]?=\s*["'][^"']{4,}["']`),
	regexp.MustCompile(`(?i)sk-[a-zA-Z0-9]{10,}`),
	regexp.MustCompile(`(?i)Bearer\s+[a-zA-Z0-9\-._~+/]+=*`),
	regexp.MustCompile(`(?i)-----BEGIN\s+(RSA\s)?PRIVATE\s+KEY-----`),
}

// ScrubSecrets replaces detected secret patterns with [REDACTED].
func ScrubSecrets(code string) string {
	result := code
	for _, re := range secretPatterns {
		result = re.ReplaceAllString(result, "[REDACTED]")
	}
	return result
}
