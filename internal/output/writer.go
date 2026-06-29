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

// resolvePath determines the final file path to write a summary report to:
//   - outPath == ""                     → defaultName in the current directory.
//   - outPath is an existing directory  → defaultName inside it.
//   - anything else                     → outPath itself, overwriting any
//     existing file there.
func resolvePath(defaultName, outPath string) string {
	if outPath == "" {
		return defaultName
	}
	if info, err := os.Stat(outPath); err == nil && info.IsDir() {
		return filepath.Join(outPath, defaultName)
	}
	return outPath
}

// WriteSummaryFile writes a summary report — including the parameters used
// to generate it — to a text file, and returns the path written to.
//
// outPath controls the destination: a blank string writes
// "<repo>_<date>.txt" (e.g. "GitSum_2026-06-28.txt") to the current
// directory; an existing directory gets that same default filename written
// inside it; any other path is used as-is, overwriting an existing file.
func WriteSummaryFile(repoPath, outPath string, params []Param, summary string) (string, error) {
	name, err := RepoName(repoPath)
	if err != nil {
		return "", err
	}

	defaultName := fmt.Sprintf("%s_%s.txt", name, time.Now().Format("2006-01-02"))
	path := resolvePath(defaultName, outPath)

	var b strings.Builder
	fmt.Fprintf(&b, "GitSum Summary — %s\n", name)
	fmt.Fprintf(&b, "Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintln(&b, "Parameters:")
	for _, p := range params {
		fmt.Fprintf(&b, "  %s: %s\n", p.Name, p.Value)
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, summary)

	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		return "", fmt.Errorf("writing summary file: %w", err)
	}

	return path, nil
}
