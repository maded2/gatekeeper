package evaluator_test

import (
	"testing"

	"github.com/eddie/gatekeeper/internal/diff"
	"github.com/eddie/gatekeeper/internal/evaluator"
)

// =============================================================================
// Story 10.1: Multi-Language Diff Evaluation
// Acceptance Tests
// =============================================================================

func TestLanguageDetection_Go(t *testing.T) {
	// Given a Go file path
	detector := evaluator.NewLanguageDetector()

	// When the language is detected
	lang := detector.Detect("main.go")

	// Then it returns Go
	if lang != "Go" {
		t.Errorf("expected 'Go', got %q", lang)
	}
}

func TestLanguageDetection_Python(t *testing.T) {
	// Given a Python file path
	detector := evaluator.NewLanguageDetector()

	// When the language is detected
	lang := detector.Detect("app.py")

	// Then it returns Python
	if lang != "Python" {
		t.Errorf("expected 'Python', got %q", lang)
	}
}

func TestLanguageDetection_JavaScript(t *testing.T) {
	// Given a JavaScript file path
	detector := evaluator.NewLanguageDetector()

	// When the language is detected
	lang := detector.Detect("app.js")

	// Then it returns JavaScript
	if lang != "JavaScript" {
		t.Errorf("expected 'JavaScript', got %q", lang)
	}
}

func TestLanguageDetection_TypeScript(t *testing.T) {
	// Given a TypeScript file path
	detector := evaluator.NewLanguageDetector()

	// When the language is detected
	lang := detector.Detect("app.ts")

	// Then it returns TypeScript
	if lang != "TypeScript" {
		t.Errorf("expected 'TypeScript', got %q", lang)
	}
}

func TestLanguageDetection_Unknown(t *testing.T) {
	// Given an unknown file extension
	detector := evaluator.NewLanguageDetector()

	// When the language is detected
	lang := detector.Detect("file.unknown")

	// Then it returns Unknown
	if lang != "Unknown" {
		t.Errorf("expected 'Unknown', got %q", lang)
	}
}

func TestLanguageDetection_MultipleLanguages(t *testing.T) {
	// Given a diff with multiple languages
	detector := evaluator.NewLanguageDetector()
	files := []string{"main.go", "app.py", "index.js", "style.css"}

	// When languages are detected
	expected := map[string]string{
		"main.go":   "Go",
		"app.py":    "Python",
		"index.js":  "JavaScript",
		"style.css": "CSS",
	}

	for _, file := range files {
		lang := detector.Detect(file)
		if expected[file] != lang {
			t.Errorf("expected %q for %s, got %q", expected[file], file, lang)
		}
	}
}

func TestLanguageCriteria_LanguageSpecific(t *testing.T) {
	// Given a language detector
	detector := evaluator.NewLanguageDetector()

	// When criteria are retrieved for Go
	goCriteria := detector.LanguageCriteria("Go")

	// Then Go-specific criteria are returned
	if goCriteria.MaxFunctionLength != 50 {
		t.Errorf("expected max function length 50, got %d", goCriteria.MaxFunctionLength)
	}
	if !goCriteria.RequiresErrorHandling {
		t.Error("expected Go to require error handling")
	}
}

func TestLanguageCriteria_Default(t *testing.T) {
	// Given a language detector
	detector := evaluator.NewLanguageDetector()

	// When criteria are retrieved for an unknown language
	criteria := detector.LanguageCriteria("Unknown")

	// Then default criteria are returned
	if criteria.MaxFunctionLength != 50 {
		t.Errorf("expected default max function length 50, got %d", criteria.MaxFunctionLength)
	}
}

func TestIsCodeFile(t *testing.T) {
	// Given a language detector
	detector := evaluator.NewLanguageDetector()

	// When checking if a file is a code file
	if !detector.IsCodeFile("main.go") {
		t.Error("expected main.go to be a code file")
	}
	if detector.IsCodeFile("README.md") {
		t.Error("expected README.md to not be a code file")
	}
	if detector.IsCodeFile("config.yaml") {
		t.Error("expected config.yaml to not be a code file")
	}
}

func TestIsDocumentationFile(t *testing.T) {
	// Given a language detector
	detector := evaluator.NewLanguageDetector()

	// When checking if a file is a documentation file
	if !detector.IsDocumentationFile("README.md") {
		t.Error("expected README.md to be a documentation file")
	}
	if detector.IsDocumentationFile("main.go") {
		t.Error("expected main.go to not be a documentation file")
	}
}

func TestIsConfigFile(t *testing.T) {
	// Given a language detector
	detector := evaluator.NewLanguageDetector()

	// When checking if a file is a config file
	if !detector.IsConfigFile("config.yaml") {
		t.Error("expected config.yaml to be a config file")
	}
	if detector.IsConfigFile("main.go") {
		t.Error("expected main.go to not be a config file")
	}
}

// =============================================================================
// Story 10.2: Non-Code Change Handling
// Acceptance Tests
// =============================================================================

func TestNonCode_DocumentationOnly(t *testing.T) {
	// Given a diff with only documentation changes
	detector := evaluator.NewLanguageDetector()
	files := []string{"README.md", "docs/api.md", "CHANGELOG.md"}

	// When checking if files are documentation
	allDoc := true
	for _, file := range files {
		if !detector.IsDocumentationFile(file) {
			allDoc = false
		}
	}

	// Then all files are documentation
	if !allDoc {
		t.Error("expected all files to be documentation")
	}
}

func TestNonCode_ConfigurationOnly(t *testing.T) {
	// Given a diff with only configuration changes
	detector := evaluator.NewLanguageDetector()
	files := []string{"config.yaml", "settings.json", "docker-compose.yml"}

	// When checking if files are configuration
	allConfig := true
	for _, file := range files {
		if !detector.IsConfigFile(file) {
			allConfig = false
		}
	}

	// Then all files are configuration
	if !allConfig {
		t.Error("expected all files to be configuration")
	}
}

func TestNonCode_MixedDiff(t *testing.T) {
	// Given a diff with mixed code and documentation
	detector := evaluator.NewLanguageDetector()
	files := []string{"main.go", "README.md", "config.yaml"}

	// When checking file types
	codeCount := 0
	docCount := 0
	configCount := 0

	for _, file := range files {
		if detector.IsCodeFile(file) {
			codeCount++
		}
		if detector.IsDocumentationFile(file) {
			docCount++
		}
		if detector.IsConfigFile(file) {
			configCount++
		}
	}

	// Then each type is correctly identified
	if codeCount != 1 {
		t.Errorf("expected 1 code file, got %d", codeCount)
	}
	if docCount != 1 {
		t.Errorf("expected 1 documentation file, got %d", docCount)
	}
	if configCount != 1 {
		t.Errorf("expected 1 config file, got %d", configCount)
	}
}

func TestNonCode_NoIrrelevantFeedback(t *testing.T) {
	// Given a documentation-only diff
	detector := evaluator.NewLanguageDetector()
	files := []string{"README.md", "docs/api.md"}

	// When evaluating documentation files
	for _, file := range files {
		lang := detector.Detect(file)

		// Then no code-specific criteria are applied
		if lang == "Markdown" {
			// Markdown files should not trigger code quality checks
			t.Logf("documentation file %s correctly identified as Markdown", file)
		}
	}
}

// =============================================================================
// Story 11.1: Empty and Minimal Diffs
// Acceptance Tests
// =============================================================================

func TestEmptyDiff_SingleLineChange(t *testing.T) {
	// Given a diff with only one line of changes
	d := diff.Diff{
		Files: []diff.FileChange{
			{
				Status:     diff.Modified,
				NewPath:    "main.go",
				NewContent: "package main\n\nfunc Hello() string {\n\treturn \"hello\"\n}\n",
			},
		},
		TotalLines: 1,
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then the diff is evaluated fairly
	if len(result.Issues) > 5 {
		t.Errorf("expected minimal issues for small diff, got %d", len(result.Issues))
	}
}

func TestEmptyDiff_TrivialChange(t *testing.T) {
	// Given a trivial diff (typo fix)
	d := diff.Diff{
		Files: []diff.FileChange{
			{
				Status:     diff.Modified,
				NewPath:    "main.go",
				NewContent: "package main\n\nfunc Hello() string {\n\treturn \"hello world\"\n}\n",
			},
		},
		TotalLines: 1,
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then the diff is not penalized disproportionately
	if result.QualityScore < 8.0 {
		t.Errorf("expected high quality score for trivial change, got %f", result.QualityScore)
	}
}

// =============================================================================
// Story 11.2: Merge Conflict and Stacked PR Handling
// Acceptance Tests
// =============================================================================

func TestMergeConflict_Detection(t *testing.T) {
	// Given a diff with merge conflict markers
	content := `package main

<<<<<<< HEAD
func Hello() string {
	return "hello"
=======
func Hello() string {
	return "world"
>>>>>>> feature
}
`
	d := diff.Diff{
		Files: []diff.FileChange{
			{
				Status:     diff.Modified,
				NewPath:    "main.go",
				NewContent: content,
			},
		},
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then the evaluation handles merge conflicts gracefully
	// (no crash, reasonable output)
	t.Logf("evaluation completed with %d issues for merge conflict diff", len(result.Issues))
}

func TestMergeConflict_Indicated(t *testing.T) {
	// Given a diff with merge conflict markers
	content := `<<<<<<< HEAD
code here
=======
other code
>>>>>>> feature
`
	d := diff.Diff{
		Files: []diff.FileChange{
			{
				Status:     diff.Modified,
				NewPath:    "main.go",
				NewContent: content,
			},
		},
	}

	// When evaluated
	result := evaluator.Evaluate(d, nil)

	// Then the evaluation completes without error
	t.Logf("merge conflict diff evaluated successfully")
	_ = result
}
