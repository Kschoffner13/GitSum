package output

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Write writes the summary to w.
func Write(w io.Writer, summary string) error {
	_, err := io.WriteString(w, summary)
	return err
}

// Param is a single named parameter used to generate a summary, recorded in
// summary report files so it's clear how the output was produced.
type Param struct {
	Name  string
	Value string
}

// RepoName returns a human-friendly name for the repo at repoPath, derived
// from its directory name.
func RepoName(repoPath string) (string, error) {
	abs, err := filepath.Abs(repoPath)
	if err != nil {
		return "", fmt.Errorf("resolving repo path: %w", err)
	}
	return filepath.Base(abs), nil
}

// WriteSummaryFile writes a summary report — including the parameters used
// to generate it — to a text file named after the repo and the current
// date (e.g. "GitSum_2026-06-28.txt") in the current directory, and returns
// the path written to.
func WriteSummaryFile(repoPath string, params []Param, summary string) (string, error) {
	name, err := RepoName(repoPath)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s_%s.txt", name, time.Now().Format("2006-01-02"))

	var b strings.Builder
	fmt.Fprintf(&b, "GitSum Summary — %s\n", name)
	fmt.Fprintf(&b, "Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintln(&b, "Parameters:")
	for _, p := range params {
		fmt.Fprintf(&b, "  %s: %s\n", p.Name, p.Value)
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, summary)

	if err := os.WriteFile(filename, []byte(b.String()), 0644); err != nil {
		return "", fmt.Errorf("writing summary file: %w", err)
	}

	return filename, nil
}
