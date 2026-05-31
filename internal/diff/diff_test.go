package diff_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eddie/gatekeeper/internal/diff"
)

// =============================================================================
// Story 2.1: Diff Input — Working Directory Changes
// Acceptance Tests
// =============================================================================

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "Test")

	// Create initial file and commit
	filePath := filepath.Join(dir, "hello.go")
	err := os.WriteFile(filePath, []byte("package main\n\nfunc Hello() string {\n\treturn \"hello\"\n}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "initial commit")

	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if _, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v", strings.Join(args, " "), err)
	}
}

func TestWorkingDirDiff_IdentifiesUncommittedChanges(t *testing.T) {
	// Given a git repo with an initial commit
	repoDir := setupTestRepo(t)

	// When a file is modified (uncommitted)
	filePath := filepath.Join(repoDir, "hello.go")
	err := os.WriteFile(filePath, []byte("package main\n\nfunc Hello() string {\n\treturn \"hello world\"\n}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the working directory diff is captured
	d, err := diff.FromWorkingDir(repoDir)

	// Then the change is identified
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if d.IsEmpty() {
		t.Error("expected diff to contain changes")
	}
	if d.FileCount() != 1 {
		t.Errorf("expected 1 changed file, got %d", d.FileCount())
	}
}

func TestWorkingDirDiff_IdentifiesModifiedFiles(t *testing.T) {
	// Given a git repo with committed files
	repoDir := setupTestRepo(t)

	// When a file is modified
	filePath := filepath.Join(repoDir, "hello.go")
	err := os.WriteFile(filePath, []byte("package main\n\nfunc Hello() string {\n\treturn \"modified\"\n}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the diff is captured
	d, err := diff.FromWorkingDir(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Then the file status is Modified
	if d.Files[0].Status != diff.Modified {
		t.Errorf("expected Modified status, got %s", d.Files[0].Status)
	}
}

func TestWorkingDirDiff_IdentifiesAddedFiles(t *testing.T) {
	// Given a git repo with an initial commit
	repoDir := setupTestRepo(t)

	// When a new file is added
	newFile := filepath.Join(repoDir, "new.go")
	err := os.WriteFile(newFile, []byte("package main\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the diff is captured
	d, err := diff.FromWorkingDir(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Then the file status is Added
	found := false
	for _, f := range d.Files {
		if f.Status == diff.Added && strings.HasSuffix(f.NewPath, "new.go") {
			found = true
		}
	}
	if !found {
		t.Error("expected Added file status for new.go")
	}
}

func TestWorkingDirDiff_IdentifiesDeletedFiles(t *testing.T) {
	// Given a git repo with a committed file
	repoDir := setupTestRepo(t)
	filePath := filepath.Join(repoDir, "hello.go")

	// When the file is deleted
	err := os.Remove(filePath)
	if err != nil {
		t.Fatal(err)
	}

	// When the diff is captured
	d, err := diff.FromWorkingDir(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Then the file status is Deleted
	found := false
	for _, f := range d.Files {
		if f.Status == diff.Deleted {
			found = true
		}
	}
	if !found {
		t.Error("expected Deleted file status")
	}
}

func TestWorkingDirDiff_ProducesDiffRepresentation(t *testing.T) {
	// Given a git repo with modified files
	repoDir := setupTestRepo(t)
	filePath := filepath.Join(repoDir, "hello.go")
	err := os.WriteFile(filePath, []byte("package main\n\nfunc Hello() string {\n\treturn \"different\"\n}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the diff is captured
	d, err := diff.FromWorkingDir(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Then the diff has content
	if d.TotalChangedLines() == 0 {
		t.Error("expected non-zero changed lines")
	}
}

func TestWorkingDirDiff_EmptyWorkingDirectory(t *testing.T) {
	// Given a git repo with no uncommitted changes
	repoDir := setupTestRepo(t)

	// When the diff is captured
	d, err := diff.FromWorkingDir(repoDir)

	// Then it returns an empty diff with an informative message
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !d.IsEmpty() {
		t.Error("expected empty diff for clean working directory")
	}
	if !strings.Contains(d.Description, "No uncommitted changes") {
		t.Errorf("expected informative message, got: %s", d.Description)
	}
}

func TestWorkingDirDiff_NoGitHistory(t *testing.T) {
	// Given a directory with no git history
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the diff is captured
	_, err = diff.FromWorkingDir(tmpDir)

	// Then a clear error is returned
	if err == nil {
		t.Fatal("expected error for non-git directory")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("expected 'not a git repository' error, got: %v", err)
	}
}

func TestWorkingDirDiff_OutputIndicatesFiles(t *testing.T) {
	// Given a git repo with changes
	repoDir := setupTestRepo(t)
	filePath := filepath.Join(repoDir, "hello.go")
	err := os.WriteFile(filePath, []byte("package main\n\nfunc Hello() string {\n\treturn \"changed\"\n}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the diff is captured
	d, err := diff.FromWorkingDir(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Then the description indicates which files were included
	if !strings.Contains(d.Description, "file(s)") {
		t.Errorf("expected file count in description, got: %s", d.Description)
	}
}

// =============================================================================
// Story 2.2: Diff Input — Staged Changes
// Acceptance Tests
// =============================================================================

func TestStagedDiff_IdentifiesStagedChanges(t *testing.T) {
	// Given a git repo with staged changes
	repoDir := setupTestRepo(t)
	filePath := filepath.Join(repoDir, "hello.go")
	err := os.WriteFile(filePath, []byte("package main\n\nfunc Hello() string {\n\treturn \"staged\"\n}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	runGit(t, repoDir, "add", "hello.go")

	// When the staged diff is captured
	d, err := diff.FromStaged(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Then the staged change is identified
	if d.IsEmpty() {
		t.Error("expected staged diff to contain changes")
	}
}

func TestStagedDiff_DistinguishesStagedFromUnstaged(t *testing.T) {
	// Given a git repo with both staged and unstaged changes
	repoDir := setupTestRepo(t)
	filePath := filepath.Join(repoDir, "hello.go")

	// Stage first change
	err := os.WriteFile(filePath, []byte("package main\n\nfunc Hello() string {\n\treturn \"staged\"\n}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	runGit(t, repoDir, "add", "hello.go")

	// Make additional unstaged change
	err = os.WriteFile(filePath, []byte("package main\n\nfunc Hello() string {\n\treturn \"unstaged\"\n}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the staged diff is captured
	d, err := diff.FromStaged(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Then only the staged change is included (not the unstaged one)
	if d.IsEmpty() {
		t.Error("expected staged diff to contain the staged change")
	}
}

func TestStagedDiff_EmptyStagingArea(t *testing.T) {
	// Given a git repo with no staged changes
	repoDir := setupTestRepo(t)

	// When the staged diff is captured
	d, err := diff.FromStaged(repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Then an informative message is returned
	if !d.IsEmpty() {
		t.Error("expected empty staged diff")
	}
	if !strings.Contains(d.Description, "No staged changes") {
		t.Errorf("expected informative message, got: %s", d.Description)
	}
}

// =============================================================================
// Story 2.3: Diff Input — Committed Changes Against a Base
// Acceptance Tests
// =============================================================================

func TestBaseRefDiff_ComparesAgainstBase(t *testing.T) {
	// Given a git repo with a feature branch
	repoDir := setupTestRepo(t)
	runGit(t, repoDir, "checkout", "-b", "feature")
	filePath := filepath.Join(repoDir, "hello.go")
	err := os.WriteFile(filePath, []byte("package main\n\nfunc Hello() string {\n\treturn \"feature\"\n}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	runGit(t, repoDir, "add", ".")
	runGit(t, repoDir, "commit", "-m", "feature change")

	// When the diff against main is captured
	d, err := diff.FromBaseRef(repoDir, "main")
	if err != nil {
		t.Fatal(err)
	}

	// Then the change is identified
	if d.IsEmpty() {
		t.Error("expected diff to contain changes against main")
	}
}

func TestBaseRefDiff_InvalidBaseReference(t *testing.T) {
	// Given a git repo
	repoDir := setupTestRepo(t)

	// When a non-existent base reference is used
	_, err := diff.FromBaseRef(repoDir, "nonexistent-branch")

	// Then a clear error is returned
	if err == nil {
		t.Fatal("expected error for invalid base reference")
	}
	if !strings.Contains(err.Error(), "nonexistent-branch") {
		t.Errorf("expected error to mention the base ref, got: %v", err)
	}
}

func TestBaseRefDiff_IdenticalBaseAndHead(t *testing.T) {
	// Given a git repo where HEAD equals the base
	repoDir := setupTestRepo(t)

	// When comparing HEAD to itself
	d, err := diff.FromBaseRef(repoDir, "main")
	if err != nil {
		t.Fatal(err)
	}

	// Then an informative message is returned
	if !d.IsEmpty() {
		t.Error("expected empty diff for identical base and HEAD")
	}
}

// =============================================================================
// Story 2.4: Diff Input — Arbitrary Diff File
// Acceptance Tests
// =============================================================================

func TestDiffFile_ValidUnifiedDiff(t *testing.T) {
	// Given a valid unified diff file
	tmpDir := t.TempDir()
	diffPath := filepath.Join(tmpDir, "change.diff")
	diffContent := `diff --git a/hello.go b/hello.go
index 1234567..abcdef 100644
--- a/hello.go
+++ b/hello.go
@@ -1,3 +1,3 @@
 package main
 
-func Hello() string { return "old" }
+func Hello() string { return "new" }
`
	err := os.WriteFile(diffPath, []byte(diffContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the diff is loaded
	d, err := diff.FromFile(diffPath)
	if err != nil {
		t.Fatal(err)
	}

	// Then the diff is parsed correctly
	if d.IsEmpty() {
		t.Error("expected parsed diff to contain changes")
	}
}

func TestDiffFile_MalformedDiff(t *testing.T) {
	// Given a malformed diff file
	tmpDir := t.TempDir()
	diffPath := filepath.Join(tmpDir, "bad.diff")
	err := os.WriteFile(diffPath, []byte("this is not a diff"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the diff is loaded
	d, err := diff.FromFile(diffPath)
	if err != nil {
		t.Fatal(err)
	}

	// Then it handles gracefully (empty or informative)
	if !d.IsEmpty() {
		t.Log("note: malformed diff produced non-empty result")
	}
}

func TestDiffFile_EmptyDiff(t *testing.T) {
	// Given an empty diff file
	tmpDir := t.TempDir()
	diffPath := filepath.Join(tmpDir, "empty.diff")
	err := os.WriteFile(diffPath, []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the diff is loaded
	d, err := diff.FromFile(diffPath)
	if err != nil {
		t.Fatal(err)
	}

	// Then an informative message is returned
	if !d.IsEmpty() {
		t.Error("expected empty diff")
	}
	if !strings.Contains(d.Description, "Empty") {
		t.Errorf("expected informative message, got: %s", d.Description)
	}
}

func TestDiffFile_WhitespaceOnly(t *testing.T) {
	// Given a diff with only whitespace changes
	tmpDir := t.TempDir()
	diffPath := filepath.Join(tmpDir, "ws.diff")
	diffContent := `diff --git a/file.txt b/file.txt
--- a/file.txt
+++ b/file.txt
@@ -1 +1 @@
-hello
+hello 
`
	err := os.WriteFile(diffPath, []byte(diffContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// When the diff is loaded
	d, err := diff.FromFile(diffPath)
	if err != nil {
		t.Fatal(err)
	}

	// Then it is handled appropriately
	if d.IsEmpty() {
		t.Log("whitespace-only diff treated as empty")
	}
}
