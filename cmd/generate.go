package cmd

import (
	"fmt"
	"os"

	"github.com/Kschoffner13/GitSum/internal/ai"
	"github.com/Kschoffner13/GitSum/internal/git"
	"github.com/Kschoffner13/GitSum/internal/output"
	"github.com/spf13/cobra"
)

var generateDays int

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a summary of the git repository",
	Long:  `Generate a summary of the git repository, including recent commit history.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			commits []git.Commit
			err     error
		)
		if generateDays > 0 {
			commits, err = git.ParseDays(".", generateDays)
		} else {
			commits, err = git.Parse(".", 20)
		}
		if err != nil {
			return fmt.Errorf("parsing repo: %w", err)
		}

		summary, err := ai.Summarize(commits)
		if err != nil {
			return fmt.Errorf("summarizing: %w", err)
		}

		return output.Write(os.Stdout, summary)
	},
}

func init() {
	generateCmd.Flags().IntVarP(&generateDays, "days", "d", 0,
		"Limit to commits from the last N days (overrides the default last-20 window)")

	rootCmd.AddCommand(generateCmd)
}
