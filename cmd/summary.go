package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Kschoffner13/GitSum/internal/agents"
	"github.com/Kschoffner13/GitSum/internal/git"
	"github.com/Kschoffner13/GitSum/internal/output"
	"github.com/spf13/cobra"
)

var (
	summaryAudience string
	summarySince    string
	summaryDays     int
	summaryLimit    int
	summaryVerbose  bool
	summaryOut      string
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Generate an AI-powered multi-perspective summary of recent commits",
	Long: `Runs a multi-agent analysis pipeline over your git history and produces
a summary tailored to your chosen audience.

Four specialist AI agents analyse the commits in parallel (technical, impact,
risk, velocity) and a synthesizer agent combines their findings into a single,
audience-appropriate output.

Audiences:
  lead-engineer   Technical detail + risk focus (default)
  manager         Outcome + team health focus, no code specifics
  client          Plain-language, benefit-framed update, zero jargon
  release-notes   Structured markdown changelog (Added / Changed / Fixed …)

Examples:
  gitsum summary
  gitsum summary --audience client
  gitsum summary --audience manager --since v1.2.0
  gitsum summary --audience release-notes --since HEAD~50
  gitsum summary --days 7
  gitsum summary --limit 40 --verbose
  gitsum summary --out report.txt
  gitsum summary --out ./reports/`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate --audience
		audience := agents.Audience(summaryAudience)
		validAudience := false
		for _, a := range agents.ValidAudiences {
			if a == audience {
				validAudience = true
				break
			}
		}
		if !validAudience {
			names := make([]string, len(agents.ValidAudiences))
			for i, a := range agents.ValidAudiences {
				names[i] = string(a)
			}
			return fmt.Errorf("unknown audience %q — valid options: %s", summaryAudience, strings.Join(names, ", "))
		}

		// Resolve API key
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY environment variable is not set")
		}

		// Parse commits
		var (
			gitCommits []git.Commit
			err        error
		)
		switch {
		case summarySince != "":
			gitCommits, err = git.ParseSince(".", summarySince)
		case summaryDays > 0:
			gitCommits, err = git.ParseDays(".", summaryDays)
		default:
			gitCommits, err = git.Parse(".", summaryLimit)
		}
		if err != nil {
			return fmt.Errorf("reading git history: %w", err)
		}
		if len(gitCommits) == 0 {
			return fmt.Errorf("no commits found (since: %q, days: %d, limit: %d)", summarySince, summaryDays, summaryLimit)
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "Analysing %d commits for audience: %s\n\n", len(gitCommits), audience)

		// Convert to agents.Commit
		commits := agents.CommitsFromGit(gitCommits)

		// Run the pipeline
		result, err := agents.Run(
			context.Background(),
			commits,
			agents.PipelineOptions{
				APIKey:   apiKey,
				Audience: audience,
				Verbose:  summaryVerbose,
			},
			func(msg string) {
				// Progress goes to stderr so stdout stays clean for piping / redirection
				fmt.Fprintln(cmd.ErrOrStderr(), msg)
			},
		)
		if err != nil {
			return err
		}

		// --verbose: print raw analyst reports before the final summary
		if summaryVerbose && len(result.Reports) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "=== ANALYST REPORTS ===\n")
			for _, r := range result.Reports {
				if r.Err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "--- %s [FAILED: %s] ---\n\n", r.Role, r.Err)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "--- %s ---\n%s\n\n", r.Role, r.Findings)
				}
			}
			fmt.Fprintln(cmd.OutOrStdout(), "=== FINAL SUMMARY ===\n")
		}

		fmt.Fprintln(cmd.OutOrStdout(), result.Summary)

		params := []output.Param{
			{Name: "audience", Value: string(audience)},
			{Name: "since", Value: summarySince},
			{Name: "days", Value: fmt.Sprintf("%d", summaryDays)},
			{Name: "limit", Value: fmt.Sprintf("%d", summaryLimit)},
			{Name: "verbose", Value: fmt.Sprintf("%v", summaryVerbose)},
		}

		path, err := output.WriteSummaryFile(".", summaryOut, params, result.Summary)
		if err != nil {
			return fmt.Errorf("writing summary file: %w", err)
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "\nSummary written to %s\n", path)

		return nil
	},
}

func init() {
	summaryCmd.Flags().StringVarP(&summaryAudience, "audience", "a", "lead-engineer",
		"Target audience: lead-engineer, manager, client, release-notes")
	summaryCmd.Flags().StringVar(&summarySince, "since", "",
		"Limit to commits after this ref/tag/date (e.g. v1.2.0, HEAD~20). Overrides --days and --limit.")
	summaryCmd.Flags().IntVarP(&summaryDays, "days", "d", 0,
		"Limit to commits from the last N days. Overrides --limit; ignored when --since is set.")
	summaryCmd.Flags().IntVarP(&summaryLimit, "limit", "n", 20,
		"Maximum number of recent commits to analyse (ignored when --since or --days is set)")
	summaryCmd.Flags().BoolVarP(&summaryVerbose, "verbose", "v", false,
		"Print raw analyst reports before the final summary")
	summaryCmd.Flags().StringVarP(&summaryOut, "out", "o", "",
		"Write the report to this file (overwritten if it exists) or, if a directory, into it as <repo>_<date>.txt")

	rootCmd.AddCommand(summaryCmd)
}
