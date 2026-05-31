package diff

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FromWorkingDir captures uncommitted changes in the working directory.
func FromWorkingDir(repoRoot string) (Diff, error) {
	if !isGitRepo(repoRoot) {
		return Diff{}, fmt.Errorf("directory %q is not a git repository", repoRoot)
	}

	// Check for uncommitted changes
	statusOutput, err := runGit(repoRoot, "status", "--porcelain")
	if err != nil {
		return Diff{}, fmt.Errorf("checking git status: %w", err)
	}

	if strings.TrimSpace(statusOutput) == "" {
		return Diff{Description: "No uncommitted changes in working directory"}, nil
	}

	// Get diff of tracked changes
	diffOutput, err := runGit(repoRoot, "diff", "--unified=3")
	if err != nil {
		return Diff{}, fmt.Errorf("generating diff: %w", err)
	}

	// Get untracked (new) files
	untrackedOutput, _ := runGit(repoRoot, "ls-files", "--others", "--exclude-standard")

	files, totalLines, err := parseUnifiedDiff(diffOutput)
	if err != nil {
		return Diff{}, fmt.Errorf("parsing diff: %w", err)
	}

	// Add untracked files
	for _, line := range strings.Split(untrackedOutput, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		files = append(files, FileChange{
			Status:  Added,
			NewPath: line,
		})
		totalLines++ // rough count for new files
	}

	if len(files) == 0 {
		return Diff{Description: "No uncommitted changes in working directory"}, nil
	}

	return Diff{
		Files:       files,
		TotalLines:  totalLines,
		Description: fmt.Sprintf("Working directory changes: %d file(s), %d line(s)", len(files), totalLines),
	}, nil
}

// FromStaged captures staged (index) changes.
func FromStaged(repoRoot string) (Diff, error) {
	if !isGitRepo(repoRoot) {
		return Diff{}, fmt.Errorf("directory %q is not a git repository", repoRoot)
	}

	// Check for staged changes using name-status
	statusOutput, err := runGit(repoRoot, "diff", "--cached", "--name-status")
	if err != nil {
		return Diff{}, fmt.Errorf("checking staged changes: %w", err)
	}

	if strings.TrimSpace(statusOutput) == "" {
		return Diff{Description: "No staged changes"}, nil
	}

	diffOutput, err := runGit(repoRoot, "diff", "--cached", "--unified=3")
	if err != nil {
		return Diff{}, fmt.Errorf("generating staged diff: %w", err)
	}

	files, totalLines, err := parseUnifiedDiff(diffOutput)
	if err != nil {
		return Diff{}, fmt.Errorf("parsing staged diff: %w", err)
	}

	return Diff{
		Files:       files,
		TotalLines:  totalLines,
		Description: fmt.Sprintf("Staged changes: %d file(s), %d line(s)", len(files), totalLines),
	}, nil
}

// FromBaseRef captures changes between HEAD and a base reference.
func FromBaseRef(repoRoot string, baseRef string) (Diff, error) {
	if !isGitRepo(repoRoot) {
		return Diff{}, fmt.Errorf("directory %q is not a git repository", repoRoot)
	}

	// Verify base ref exists
	_, err := runGit(repoRoot, "rev-parse", "--verify", baseRef)
	if err != nil {
		return Diff{}, fmt.Errorf("base reference %q does not exist: %w", baseRef, err)
	}

	diffOutput, err := runGit(repoRoot, "diff", "--unified=3", baseRef+"...HEAD")
	if err != nil {
		return Diff{}, fmt.Errorf("generating diff against %s: %w", baseRef, err)
	}

	if strings.TrimSpace(diffOutput) == "" {
		return Diff{Description: fmt.Sprintf("No differences between HEAD and %s", baseRef)}, nil
	}

	files, totalLines, err := parseUnifiedDiff(diffOutput)
	if err != nil {
		return Diff{}, fmt.Errorf("parsing diff: %w", err)
	}

	return Diff{
		Files:       files,
		TotalLines:  totalLines,
		Description: fmt.Sprintf("Changes against %s: %d file(s), %d line(s)", baseRef, len(files), totalLines),
	}, nil
}

// FromFile reads a diff from a file path.
func FromFile(path string) (Diff, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Diff{}, fmt.Errorf("reading diff file %q: %w", path, err)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return Diff{Description: "Empty diff file"}, nil
	}

	files, totalLines, err := parseUnifiedDiff(string(data))
	if err != nil {
		return Diff{}, fmt.Errorf("parsing diff file: %w", err)
	}

	if len(files) == 0 {
		return Diff{Description: "No meaningful code changes in diff file (whitespace-only)"}, nil
	}

	return Diff{
		Files:       files,
		TotalLines:  totalLines,
		Description: fmt.Sprintf("Diff from file: %d file(s), %d line(s)", len(files), totalLines),
	}, nil
}

// FromReader reads a diff from a reader (stdin).
func FromReader(data []byte) (Diff, error) {
	if len(strings.TrimSpace(string(data))) == 0 {
		return Diff{Description: "Empty diff input"}, nil
	}

	files, totalLines, err := parseUnifiedDiff(string(data))
	if err != nil {
		return Diff{}, fmt.Errorf("parsing diff input: %w", err)
	}

	return Diff{
		Files:       files,
		TotalLines:  totalLines,
		Description: fmt.Sprintf("Diff from input: %d file(s), %d line(s)", len(files), totalLines),
	}, nil
}

func isGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, output)
	}
	return string(output), nil
}

func parseUnifiedDiff(diffOutput string) ([]FileChange, int, error) {
	var files []FileChange
	totalLines := 0
	lines := strings.Split(diffOutput, "\n")

	var currentFile FileChange
	inHeader := true

	for _, line := range lines {
		// File header
		if strings.HasPrefix(line, "diff --git") {
			if currentFile.NewPath != "" || currentFile.OldPath != "" {
				files = append(files, currentFile)
			}
			currentFile = FileChange{}
			inHeader = true
			continue
		}

		// Old/new file paths
		if inHeader && strings.HasPrefix(line, "--- ") {
			path := strings.TrimPrefix(line, "--- ")
			if !strings.HasPrefix(path, "a/") {
				currentFile.OldPath = path
			} else {
				currentFile.OldPath = strings.TrimPrefix(path, "a/")
			}
			continue
		}
		if inHeader && strings.HasPrefix(line, "+++ ") {
			path := strings.TrimPrefix(line, "+++ ")
			if !strings.HasPrefix(path, "b/") {
				currentFile.NewPath = path
			} else {
				currentFile.NewPath = strings.TrimPrefix(path, "b/")
			}
			// Determine status
			if currentFile.OldPath == "/dev/null" {
				currentFile.Status = Added
			} else if currentFile.NewPath == "/dev/null" {
				currentFile.Status = Deleted
			} else {
				currentFile.Status = Modified
			}
			inHeader = false
			continue
		}

		// Count changed lines
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			totalLines++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			totalLines++
		}
	}

	if currentFile.NewPath != "" || currentFile.OldPath != "" {
		files = append(files, currentFile)
	}

	return files, totalLines, nil
}
