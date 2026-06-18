package cmd

import (
	"fmt"
	"os"

	"github.com/Kschoffner13/GitSum/internal/ai"
	"github.com/Kschoffner13/GitSum/internal/git"
	"github.com/Kschoffner13/GitSum/internal/output"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a summary of the git repository",
	Long:  `Generate a summary of the git repository, including recent commit history.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		commits, err := git.Parse(".", 20)
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
	rootCmd.AddCommand(generateCmd)
}
