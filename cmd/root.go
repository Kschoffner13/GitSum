/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/Kschoffner13/GitSum/internal/config"
	"github.com/Kschoffner13/GitSum/internal/version"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "gitsum",
	Short: "AI-powered git history summarizer",
	Long: `GitSum analyses your git commit history using AI and produces summaries
tailored to different audiences — engineers, managers, clients, and changelogs.

Examples:
  gitsum generate                              # plain commit list
  gitsum summary                               # full analysis, lead-engineer view
  gitsum summary --audience client             # plain-language client update
  gitsum summary --audience release-notes --since v1.2.0`,
	Version: version.Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return config.Load()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
