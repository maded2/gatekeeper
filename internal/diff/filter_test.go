package diff_test

import (
	"testing"

	"github.com/eddie/gatekeeper/internal/diff"
)

// =============================================================================
// Story 2.5: Diff Filtering
// Acceptance Tests
// =============================================================================

func TestFilter_ExcludesLockFiles(t *testing.T) {
	d := diff.Diff{
		Files: []diff.FileChange{
			{Status: diff.Modified, NewPath: "package-lock.json"},
			{Status: diff.Modified, NewPath: "yarn.lock"},
			{Status: diff.Modified, NewPath: "src/main.go"},
		},
	}

	filtered, excluded := diff.Filter(d, []string{"*.lock", "package-lock.json"})

	if len(filtered.Files) != 1 {
		t.Errorf("expected 1 file after filtering, got %d", len(filtered.Files))
	}
	if filtered.Files[0].NewPath != "src/main.go" {
		t.Errorf("expected src/main.go to remain, got %s", filtered.Files[0].NewPath)
	}
	if len(excluded) < 1 {
		t.Error("expected excluded files to be reported")
	}
}

func TestFilter_ExcludesGeneratedCode(t *testing.T) {
	d := diff.Diff{
		Files: []diff.FileChange{
			{Status: diff.Modified, NewPath: "proto/user.pb.go"},
			{Status: diff.Modified, NewPath: "src/main.go"},
		},
	}

	filtered, _ := diff.Filter(d, []string{"*.pb.go", "*_gen.go"})

	if len(filtered.Files) != 1 {
		t.Errorf("expected 1 file after filtering, got %d", len(filtered.Files))
	}
}

func TestFilter_ExcludesBinaryAssets(t *testing.T) {
	d := diff.Diff{
		Files: []diff.FileChange{
			{Status: diff.Modified, NewPath: "image.png"},
			{Status: diff.Modified, NewPath: "style.min.css"},
			{Status: diff.Modified, NewPath: "src/main.go"},
		},
	}

	filtered, _ := diff.Filter(d, []string{"*.png", "*.min.*"})

	if len(filtered.Files) != 1 {
		t.Errorf("expected 1 file after filtering, got %d", len(filtered.Files))
	}
}

func TestFilter_CustomPatterns(t *testing.T) {
	d := diff.Diff{
		Files: []diff.FileChange{
			{Status: diff.Modified, NewPath: "vendor/github.com/pkg/file.go"},
			{Status: diff.Modified, NewPath: "src/main.go"},
		},
	}

	filtered, _ := diff.Filter(d, []string{"vendor/*"})

	if len(filtered.Files) != 1 {
		t.Errorf("expected 1 file after filtering, got %d", len(filtered.Files))
	}
}

func TestFilter_ExcludedFilesReported(t *testing.T) {
	d := diff.Diff{
		Files: []diff.FileChange{
			{Status: diff.Modified, NewPath: "package-lock.json"},
			{Status: diff.Modified, NewPath: "src/main.go"},
		},
	}

	_, excluded := diff.Filter(d, []string{"*.lock", "package-lock.json"})

	if len(excluded) == 0 {
		t.Error("expected excluded files to be reported")
	}
}

func TestFilter_NoFalseExclusions(t *testing.T) {
	d := diff.Diff{
		Files: []diff.FileChange{
			{Status: diff.Modified, NewPath: "src/main.go"},
			{Status: diff.Modified, NewPath: "src/utils.go"},
		},
	}

	filtered, _ := diff.Filter(d, []string{"*.lock"})

	if len(filtered.Files) != 2 {
		t.Errorf("expected 2 files to remain, got %d", len(filtered.Files))
	}
}
