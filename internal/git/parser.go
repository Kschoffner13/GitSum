// Package git provides utilities for reading and parsing git commit history.
package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Commit represents a single git commit.
type Commit struct {
	Hash      string
	Author    string
	Timestamp string // ISO 8601 author date
	Message   string
	Diff      string // --stat summary of files changed in this commit
}

// Parse returns the most recent n commits from the repository at repoPath.
//
// It enriches each commit with a --stat diff summary so downstream consumers
// (e.g. the agents pipeline) can reason about which files changed without
// incurring the full token cost of a unified diff.
func Parse(repoPath string, limit int) ([]Commit, error) {
	args := []string{"-C", repoPath, "log", fmt.Sprintf("-n%d", limit), "--pretty=format:%H|%an|%ai|%s"}
	return runGitLog(repoPath, args)
}

// ParseSince returns all commits reachable from HEAD but not from the given
// ref (a branch name, tag, or date string accepted by git-log --after).
//
// Example refs: "v1.2.0", "HEAD~20", "2024-01-01"
func ParseSince(repoPath, since string) ([]Commit, error) {
	args := []string{"-C", repoPath, "log", since + "..HEAD", "--pretty=format:%H|%an|%ai|%s"}
	return runGitLog(repoPath, args)
}

// runGitLog executes a git-log command with the given args and parses stdout
// into a []Commit slice. Each commit is enriched with a --stat diff.
func runGitLog(repoPath string, args []string) ([]Commit, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("running git log: %w", err)
	}

	var commits []Commit
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}

		// Format: HASH|AUTHOR|TIMESTAMP|MESSAGE
		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			continue
		}

		hash := parts[0]
		diff, _ := commitStat(repoPath, hash) // best-effort; ignore errors

		commits = append(commits, Commit{
			Hash:      hash,
			Author:    parts[1],
			Timestamp: parts[2],
			Message:   parts[3],
			Diff:      diff,
		})
	}

	return commits, nil
}

// commitStat returns a --stat summary for a single commit.
//
// Using --stat rather than a full unified diff keeps prompt token usage
// manageable for large histories. Switch to "show --unified=3 <hash>" if
// you need line-level diff content for a focused range.
func commitStat(repoPath, hash string) (string, error) {
	out, err := exec.Command("git", "-C", repoPath, "show", "--stat", "--no-color", hash).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
