package evaluator

import (
	"path/filepath"
	"strings"
)

// LanguageDetector identifies the programming language of a file.
type LanguageDetector struct {
	extensions map[string]string
}

// NewLanguageDetector creates a new language detector with known language mappings.
func NewLanguageDetector() *LanguageDetector {
	extensions := map[string]string{
		".go":    "Go",
		".py":    "Python",
		".js":    "JavaScript",
		".ts":    "TypeScript",
		".tsx":   "TypeScript",
		".jsx":   "JavaScript",
		".java":  "Java",
		".kt":    "Kotlin",
		".rs":    "Rust",
		".c":     "C",
		".cpp":   "C++",
		".cc":    "C++",
		".h":     "C/C++",
		".hpp":   "C++",
		".rb":    "Ruby",
		".php":   "PHP",
		".swift": "Swift",
		".scala": "Scala",
		".rkt":   "Racket",
		".clj":   "Clojure",
		".hs":    "Haskell",
		".erl":   "Erlang",
		".ex":    "Elixir",
		".exs":   "Elixir",
		".lua":   "Lua",
		".sh":    "Shell",
		".bash":  "Bash",
		".zsh":   "Zsh",
		".sql":   "SQL",
		".html":  "HTML",
		".css":   "CSS",
		".scss":  "SCSS",
		".less":  "Less",
		".yaml":  "YAML",
		".yml":   "YAML",
		".json":  "JSON",
		".xml":   "XML",
		".toml":  "TOML",
		".ini":   "INI",
		".cfg":   "Config",
		".conf":  "Config",
		".md":    "Markdown",
		".txt":   "Text",
		".rst":   "reStructuredText",
		".adoc":  "AsciiDoc",
		".dockerfile": "Dockerfile",
		".tf":    "Terraform",
		".hcl":   "HCL",
		".proto": "Protocol Buffers",
		".graphql": "GraphQL",
		".vim":   "VimScript",
		".ps1":   "PowerShell",
		".psm1":  "PowerShell",
		".psd1":  "PowerShell",
		".bat":   "Batch",
		".cmd":   "Batch",
	}

	return &LanguageDetector{extensions: extensions}
}

// Detect identifies the language of a file based on its extension.
func (ld *LanguageDetector) Detect(filePath string) string {
	ext := filepath.Ext(filePath)
	if lang, ok := ld.extensions[ext]; ok {
		return lang
	}

	// Handle special cases
	if strings.HasSuffix(filePath, "Dockerfile") {
		return "Dockerfile"
	}
	if strings.HasSuffix(filePath, "Makefile") {
		return "Makefile"
	}
	if strings.HasSuffix(filePath, "Rakefile") {
		return "Ruby"
	}
	if strings.HasSuffix(filePath, "Gemfile") {
		return "Ruby"
	}
	if strings.HasSuffix(filePath, "Vagrantfile") {
		return "Ruby"
	}
	if strings.HasPrefix(filepath.Base(filePath), ".") && strings.HasSuffix(filePath, "bash") {
		return "Bash"
	}
	if strings.HasPrefix(filepath.Base(filePath), ".") && strings.HasSuffix(filePath, "zsh") {
		return "Zsh"
	}

	return "Unknown"
}

// IsCodeFile returns true if the file is a code file (not documentation or config).
func (ld *LanguageDetector) IsCodeFile(filePath string) bool {
	lang := ld.Detect(filePath)
	nonCodeLanguages := map[string]bool{
		"Markdown":          true,
		"Text":              true,
		"reStructuredText":  true,
		"AsciiDoc":          true,
		"YAML":              true,
		"JSON":              true,
		"XML":               true,
		"TOML":              true,
		"INI":               true,
		"Config":            true,
		"HTML":              true,
		"CSS":               true,
		"SCSS":              true,
		"Less":              true,
	}

	return !nonCodeLanguages[lang]
}

// IsDocumentationFile returns true if the file is a documentation file.
func (ld *LanguageDetector) IsDocumentationFile(filePath string) bool {
	lang := ld.Detect(filePath)
	docLanguages := map[string]bool{
		"Markdown":          true,
		"Text":              true,
		"reStructuredText":  true,
		"AsciiDoc":          true,
		"HTML":              true,
	}

	return docLanguages[lang]
}

// IsConfigFile returns true if the file is a configuration file.
func (ld *LanguageDetector) IsConfigFile(filePath string) bool {
	lang := ld.Detect(filePath)
	configLanguages := map[string]bool{
		"YAML":   true,
		"JSON":   true,
		"XML":    true,
		"TOML":   true,
		"INI":    true,
		"Config": true,
		"HCL":    true,
	}

	return configLanguages[lang]
}

// LanguageCriteria returns evaluation criteria specific to a language.
func (ld *LanguageDetector) LanguageCriteria(lang string) LanguageCriteria {
	switch lang {
	case "Go":
		return LanguageCriteria{
			MaxFunctionLength:     50,
			MaxNestingDepth:       4,
			MaxParameterCount:     5,
			RequiresErrorHandling: true,
			HasMainFunction:       true,
		}
	case "Python":
		return LanguageCriteria{
			MaxFunctionLength:     50,
			MaxNestingDepth:       4,
			MaxParameterCount:     5,
			RequiresErrorHandling: true,
			HasMainFunction:       false,
		}
	case "JavaScript", "TypeScript":
		return LanguageCriteria{
			MaxFunctionLength:     50,
			MaxNestingDepth:       4,
			MaxParameterCount:     5,
			RequiresErrorHandling: true,
			HasMainFunction:       false,
		}
	case "Java", "Kotlin":
		return LanguageCriteria{
			MaxFunctionLength:     50,
			MaxNestingDepth:       4,
			MaxParameterCount:     5,
			RequiresErrorHandling: true,
			HasMainFunction:       true,
		}
	case "Rust":
		return LanguageCriteria{
			MaxFunctionLength:     50,
			MaxNestingDepth:       4,
			MaxParameterCount:     5,
			RequiresErrorHandling: true,
			HasMainFunction:       true,
		}
	default:
		return LanguageCriteria{
			MaxFunctionLength:     50,
			MaxNestingDepth:       4,
			MaxParameterCount:     5,
			RequiresErrorHandling: true,
			HasMainFunction:       false,
		}
	}
}

// LanguageCriteria defines evaluation criteria for a specific language.
type LanguageCriteria struct {
	MaxFunctionLength     int
	MaxNestingDepth       int
	MaxParameterCount     int
	RequiresErrorHandling bool
	HasMainFunction       bool
}
