// Package agents implements a multi-agent pipeline that analyses git commit
// history from several specialised perspectives and synthesises the results
// into an audience-appropriate summary.
//
// # Architecture
//
// The pipeline runs in two stages:
//
//  1. Parallel analyst agents — four goroutines each examine the same commit
//     data through a different lens (technical, impact, risk, velocity) and
//     return a structured [AnalystReport].
//
//  2. Synthesizer — receives all analyst reports and a target [Audience] and
//     produces a single, appropriately framed summary.
//
// Entry point: [Run].
package agents

// Commit is the unit of input for the pipeline.
// It mirrors [git.Commit] but is defined here so the agents package has no
// import cycle with the git package.
type Commit struct {
	Hash      string
	Author    string
	Timestamp string
	Message   string
	Diff      string // --stat summary; may be empty
}

// AnalystRole identifies which analytical perspective a report represents.
type AnalystRole string

const (
	// RoleTechnical focuses on code-level changes: files, modules, patterns.
	RoleTechnical AnalystRole = "technical"

	// RoleImpact focuses on functional/product changes visible to users or
	// downstream consumers.
	RoleImpact AnalystRole = "impact"

	// RoleRisk surfaces breaking changes, missing tests, migration steps, and
	// anything that could affect deployment stability.
	RoleRisk AnalystRole = "risk"

	// RoleVelocity characterises how the work was done: cadence, author
	// distribution, churn, and commit quality.
	RoleVelocity AnalystRole = "velocity"
)

// AnalystReport is the structured output from one analyst agent.
// If the API call failed, Err is non-nil and Findings is empty.
type AnalystReport struct {
	Role     AnalystRole
	Findings string
	Err      error
}

// Audience controls how the synthesizer frames its final output.
type Audience string

const (
	// AudienceLeadEngineer emphasises technical detail and risk.
	// Tone: precise, dense, peer-to-peer.
	AudienceLeadEngineer Audience = "lead-engineer"

	// AudienceManager emphasises outcomes and team health.
	// Tone: outcome-focused, no implementation detail.
	AudienceManager Audience = "manager"

	// AudienceClient uses plain language framed around user benefits.
	// Tone: professional, zero jargon.
	AudienceClient Audience = "client"

	// AudienceReleaseNotes produces a structured markdown changelog.
	// Tone: formal, versioned, public-safe.
	AudienceReleaseNotes Audience = "release-notes"
)

// ValidAudiences lists every supported audience value.
// Used by the CLI for flag validation and help text.
var ValidAudiences = []Audience{
	AudienceLeadEngineer,
	AudienceManager,
	AudienceClient,
	AudienceReleaseNotes,
}

// SummaryResult is the final output of a complete pipeline run.
type SummaryResult struct {
	// Audience is the target audience this summary was written for.
	Audience Audience

	// Summary is the synthesized, audience-appropriate markdown text.
	Summary string

	// Reports holds the raw analyst output. Populated only when
	// PipelineOptions.Verbose is true.
	Reports []AnalystReport
}
