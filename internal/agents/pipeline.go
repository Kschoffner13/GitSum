package agents

import (
	"context"
	"fmt"

	"github.com/Kschoffner13/GitSum/internal/git"
)

// PipelineOptions configures a single pipeline run.
type PipelineOptions struct {
	// APIKey is the Anthropic API key. Required.
	APIKey string

	// Audience controls how the synthesizer frames its output. Required.
	Audience Audience

	// Verbose causes SummaryResult.Reports to be populated with the raw
	// analyst findings, which the CLI prints before the final summary.
	Verbose bool
}

// ProgressFunc is an optional callback that receives human-readable status
// messages during pipeline execution. It is called from the goroutine that
// runs the pipeline, not from analyst goroutines, so it is safe to write to
// stderr without a mutex.
type ProgressFunc func(msg string)

// CommitsFromGit is a convenience adapter that converts [git.Commit] values
// (from the git package) into the [Commit] type expected by this package,
// avoiding an import cycle between agents and git.
func CommitsFromGit(gc []git.Commit) []Commit {
	out := make([]Commit, len(gc))
	for i, c := range gc {
		out[i] = Commit{
			Hash:      c.Hash,
			Author:    c.Author,
			Timestamp: c.Timestamp,
			Message:   c.Message,
			Diff:      c.Diff,
		}
	}
	return out
}

// Run executes the full analyst → synthesizer pipeline and returns a
// [SummaryResult]. progress may be nil.
//
// Partial analyst failures are tolerated — the pipeline aborts only when
// every analyst has failed, since a synthesizer with no input is useless.
func Run(ctx context.Context, commits []Commit, opts PipelineOptions, progress ProgressFunc) (SummaryResult, error) {
	if progress == nil {
		progress = func(string) {}
	}

	if len(commits) == 0 {
		return SummaryResult{}, fmt.Errorf("no commits to analyse")
	}

	// Stage 1: run all analysts in parallel
	progress("Running analyst agents in parallel…")
	reports := RunAnalysts(ctx, opts.APIKey, commits)

	failCount := 0
	for _, r := range reports {
		if r.Err != nil {
			failCount++
			progress(fmt.Sprintf("  ⚠  %s analyst failed: %s", r.Role, r.Err))
		} else {
			progress(fmt.Sprintf("  ✓  %s analyst complete", r.Role))
		}
	}

	if failCount == len(reports) {
		return SummaryResult{}, fmt.Errorf("all analyst agents failed — cannot synthesize")
	}

	// Stage 2: synthesize into audience-appropriate output
	progress(fmt.Sprintf("Synthesizing summary for audience: %s…", opts.Audience))
	summary, err := Synthesize(ctx, opts.APIKey, reports, opts.Audience)
	if err != nil {
		return SummaryResult{}, fmt.Errorf("synthesis failed: %w", err)
	}

	result := SummaryResult{
		Audience: opts.Audience,
		Summary:  summary,
	}
	if opts.Verbose {
		result.Reports = reports
	}

	return result, nil
}
