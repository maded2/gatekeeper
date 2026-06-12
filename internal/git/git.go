package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// CommitInfo holds metadata about a single commit.
type CommitInfo struct {
	Hash    string
	Message string
}

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

// GetCommitsInRange returns the list of commits in the given range.
func GetCommitsInRange(repoPath, rangeSpec string) ([]CommitInfo, error) {
	cmd := exec.Command("git", "log", rangeSpec, "--format=%H %s")
	cmd.Dir = repoPath

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git log %s: %w", rangeSpec, err)
	}

	var commits []CommitInfo
	re := regexp.MustCompile(`^([0-9a-f]+) (.+)$`)
	for _, line := range strings.Split(out.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if matches != nil {
			commits = append(commits, CommitInfo{
				Hash:    matches[1],
				Message: matches[2],
			})
		}
	}

	return commits, nil
}

// GetChangedFilesInRange returns the list of files changed in the commit range.
func GetChangedFilesInRange(repoPath, rangeSpec string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", rangeSpec)
	cmd.Dir = repoPath

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git diff --name-only %s: %w", rangeSpec, err)
	}

	files := strings.Fields(out.String())
	return files, nil
}

// GetDiffInRange returns the full diff for the commit range.
func GetDiffInRange(repoPath, rangeSpec string) (string, error) {
	cmd := exec.Command("git", "diff", rangeSpec)
	cmd.Dir = repoPath

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git diff %s: %w", rangeSpec, err)
	}

	return out.String(), nil
}
