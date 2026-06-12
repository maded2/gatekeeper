package git_test

import (
	"testing"

	"gatekeeper/internal/git"
)

// --- Story C-2: Check Quality of a Specific Commit Range ---

// ACCEPTANCE CRITERIA 1:
// "I can specify a commit range (e.g., HEAD~3..HEAD) to evaluate"
func TestRange_GetCommits(t *testing.T) {
	dir := setupGitRepo(t)

	commits, err := git.GetCommitsInRange(dir, "HEAD~1..HEAD")
	if err != nil {
		t.Fatalf("GetCommitsInRange returned error: %v", err)
	}

	if len(commits) != 1 {
		t.Errorf("expected 1 commit in range, got %d", len(commits))
	}
}

// ACCEPTANCE CRITERIA 2:
// "Only the changes introduced in the specified commits are analyzed"
func TestRange_GetChangedFilesInRange(t *testing.T) {
	dir := setupGitRepo(t)

	files, err := git.GetChangedFilesInRange(dir, "HEAD~1..HEAD")
	if err != nil {
		t.Fatalf("GetChangedFilesInRange returned error: %v", err)
	}

	found := false
	for _, f := range files {
		if f == "main.go" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected main.go in changed files, got: %v", files)
	}
}

// ACCEPTANCE CRITERIA 3:
// "The output aggregates findings across all commits in the range"
// (Verified by evaluator integration)

// ACCEPTANCE CRITERIA 4:
// "I can see which commit introduced each finding"
func TestRange_CommitHasMessage(t *testing.T) {
	dir := setupGitRepo(t)

	commits, err := git.GetCommitsInRange(dir, "HEAD~1..HEAD")
	if err != nil {
		t.Fatalf("GetCommitsInRange returned error: %v", err)
	}

	if len(commits) == 0 {
		t.Fatal("expected commits in range")
	}

	if commits[0].Message == "" {
		t.Error("expected commit message to be populated")
	}
	if commits[0].Hash == "" {
		t.Error("expected commit hash to be populated")
	}
}
