package ai

import (
	"fmt"
	"strings"

	"github.com/Kschoffner13/GitSum/internal/git"
)

// Summarize produces a human-readable summary of the given commits.
//
// This is currently a placeholder that lists commit messages; it's the
// seam where a real AI-backed summarization call would plug in.
func Summarize(commits []git.Commit) (string, error) {
	if len(commits) == 0 {
		return "No commits to summarize.\n", nil
	}

	var b strings.Builder
	b.WriteString("Recent changes:\n")
	for _, c := range commits {
		fmt.Fprintf(&b, "- %s (%s)\n", c.Message, c.Author)
	}

	return b.String(), nil
}
