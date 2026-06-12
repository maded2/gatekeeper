package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GetChangedFiles returns the list of files that differ between base and target branches.
func GetChangedFiles(repoPath, base, target string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", fmt.Sprintf("%s...%s", base, target))
	cmd.Dir = repoPath

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git diff %s...%s: %w", base, target, err)
	}

	files := strings.Fields(out.String())
	return files, nil
}

// GetDiff returns the full diff content between base and target branches.
func GetDiff(repoPath, base, target string) (string, error) {
	cmd := exec.Command("git", "diff", fmt.Sprintf("%s...%s", base, target))
	cmd.Dir = repoPath

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git diff %s...%s: %w", base, target, err)
	}

	return out.String(), nil
}

// GetFileDiff returns the diff for a single file between base and target branches.
func GetFileDiff(repoPath, base, target, file string) (string, error) {
	cmd := exec.Command("git", "diff", fmt.Sprintf("%s...%s", base, target), "--", file)
	cmd.Dir = repoPath

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git diff %s...%s -- %s: %w", base, target, file, err)
	}

	return out.String(), nil
}
