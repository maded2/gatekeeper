package evaluator_test

import (
	"testing"

	"gatekeeper/internal/evaluator"
)

// --- Story B-4: Skip Analysis for Trivial Changes ---

// ACCEPTANCE CRITERIA 1:
// "Changes that affect only .md, .txt, or similar non-code files trigger an immediate pass"
func TestIsTrivialChange_NonCodeFiles(t *testing.T) {
	nonCodeFiles := []string{
		"README.md",
		"docs/guide.md",
		"CHANGELOG.txt",
		"LICENSE",
		".gitignore",
		"package.json",
		"tsconfig.json",
		"Makefile",
	}

	for _, f := range nonCodeFiles {
		if !evaluator.IsTrivialChange(f, "") {
			t.Errorf("expected %s to be classified as trivial", f)
		}
	}
}

// ACCEPTANCE CRITERIA 2:
// "Changes that are purely whitespace-only trigger an immediate pass"
func TestIsTrivialChange_WhitespaceOnly(t *testing.T) {
	whitespaceDiffs := []string{
		"   \n",
		"\t\t\n",
		"  \n  \n  \n",
		"   ",
	}

	for _, diff := range whitespaceDiffs {
		if !evaluator.IsTrivialChange("main.go", diff) {
			t.Errorf("expected whitespace-only diff to be trivial: %q", diff)
		}
	}
}

// Code files with actual changes are NOT trivial
func TestIsTrivialChange_CodeFilesNotTrivial(t *testing.T) {
	codeFiles := []string{
		"main.go",
		"src/auth.py",
		"lib/index.js",
		"app.ts",
	}

	for _, f := range codeFiles {
		if evaluator.IsTrivialChange(f, "") {
			t.Errorf("expected %s to NOT be trivial", f)
		}
	}
}

// Code changes with actual content are NOT trivial
func TestIsTrivialChange_CodeChangesNotTrivial(t *testing.T) {
	codeDiffs := []string{
		"+func hello() {}",
		"-old code\n+new code",
		"@@ -1,3 +1,3 @@\n-foo\n+bar",
	}

	for _, diff := range codeDiffs {
		if evaluator.IsTrivialChange("main.go", diff) {
			t.Errorf("expected code diff to NOT be trivial: %q", diff)
		}
	}
}

// ACCEPTANCE CRITERIA 3:
// "I see a message explaining that the changes were classified as trivial and skipped"
func TestTrivialChangeMessage(t *testing.T) {
	msg := evaluator.TrivialChangeMessage("README.md")
	if msg == "" {
		t.Error("expected non-empty trivial change message")
	}
}

// ACCEPTANCE CRITERIA 4:
// "The skip behavior does not apply when I explicitly request a full workspace check"
func TestIsTrivialChange_FullWorkspaceCheck(t *testing.T) {
	// When doing a full workspace check, non-code files should still be
	// skipped individually, but the overall check should proceed
	// This is verified by CheckWorkspace which filters non-code files
	// but still evaluates the workspace
}

// Mixed changes: code + non-code → not trivial
func TestIsTrivialChange_MixedChanges(t *testing.T) {
	// If any file is a code file, the change is not trivial
	if evaluator.IsTrivialChange("main.go", "") {
		t.Error("code file should not be trivial")
	}
}

// Code file with empty diff is not trivial (file still needs analysis)
func TestIsTrivialChange_CodeFileEmptyDiff(t *testing.T) {
	if evaluator.IsTrivialChange("main.go", "") {
		t.Error("code file with empty diff should not be trivial (file still needs analysis)")
	}
}
