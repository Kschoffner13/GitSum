package agents

import (
	"context"
	"fmt"
	"strings"
)

// audienceConfig controls the synthesizer's behaviour for a given audience.
type audienceConfig struct {
	label        string        // human-readable name used in the system prompt
	tone         string        // one-line tone description
	instructions string        // detailed writing instructions for the synthesizer
}

// audienceConfigs maps each [Audience] to its synthesizer configuration.
// To add a new audience, add an entry here and a constant in types.go.
var audienceConfigs = map[Audience]audienceConfig{
	AudienceLeadEngineer: {
		label: "Lead Engineer",
		tone:  "precise, dense, peer-to-peer",
		instructions: `Write for a lead engineer who wants the full picture fast.
- Lead with the most significant technical changes
- Call out every risk, breaking change, or migration requirement explicitly
- Include relevant implementation details where they matter
- Skip business context; keep velocity observations brief
- Use technical terminology freely
- Format: short intro paragraph, then structured sections with ## headers`,
	},
	AudienceManager: {
		label: "Engineering Manager",
		tone:  "outcome-focused, no implementation detail",
		instructions: `Write for an engineering manager who cares about outcomes and team health.
- Lead with what changed from a product/feature perspective
- Summarise delivery health: was work clean, on-track, scattered?
- Flag any risks that affect timeline, stability, or the next sprint
- Do NOT include code-level details, file names, or diff specifics
- Format: brief narrative summary (2-3 paragraphs), then a bullet-point risk/attention list`,
	},
	AudienceClient: {
		label: "Client / External Stakeholder",
		tone:  "plain language, benefit-framed, zero jargon",
		instructions: `Write for a non-technical client or external stakeholder.
- Only include what changed from the user's perspective
- Frame every change as a benefit or resolved issue ("You can now...", "We fixed an issue where...")
- Completely omit: file names, technical terms, risk details, velocity observations
- Never mention internal implementation choices
- Tone: professional but warm, like a polished product update email
- Format: 1-2 sentence intro, then a short bulleted "What's new" list`,
	},
	AudienceReleaseNotes: {
		label: "Release Notes",
		tone:  "structured, versioned, public-safe",
		instructions: `Write formal release notes suitable for a changelog or public release page.
- Use standard changelog categories: Added, Changed, Fixed, Deprecated, Removed, Security
- Only include categories that actually apply — omit empty ones
- Each item: one concise sentence, starting with a past-tense verb
- Include breaking changes in a clearly marked section at the top if any exist
- Omit internal velocity/team observations entirely
- Format: markdown with ## headers per category`,
	},
}

// buildSystemPrompt returns the synthesizer's system prompt for the given audience.
func buildSystemPrompt(audience Audience) string {
	cfg := audienceConfigs[audience]
	return fmt.Sprintf(`You are a technical writer synthesising multiple analysis reports about a git commit history.
Your audience: %s
Tone: %s

Instructions:
%s

You will receive structured reports from four specialist analysts:
- Technical Analyst (code-level changes)
- Impact Analyst    (functional/product changes)
- Risk Analyst      (risks, breaking changes, migration needs)
- Velocity Analyst  (team health, commit cadence)

Draw on all four reports but weight them according to what matters for your audience.
Do not mention the analysts or the analysis process in your output.
Produce only the final summary — nothing else.`, cfg.label, cfg.tone, cfg.instructions)
}

// buildUserPrompt assembles the analyst reports into the synthesizer's user turn.
func buildUserPrompt(reports []AnalystReport) string {
	roleLabels := map[AnalystRole]string{
		RoleTechnical: "TECHNICAL ANALYST",
		RoleImpact:    "IMPACT ANALYST",
		RoleRisk:      "RISK ANALYST",
		RoleVelocity:  "VELOCITY ANALYST",
	}

	var sb strings.Builder
	sb.WriteString("Here are the analyst reports:\n\n")

	for _, r := range reports {
		label := roleLabels[r.Role]
		fmt.Fprintf(&sb, "=== %s ===\n", label)
		if r.Err != nil {
			fmt.Fprintf(&sb, "[Analysis failed: %s]\n\n", r.Err.Error())
		} else {
			sb.WriteString(r.Findings)
			sb.WriteString("\n\n")
		}
	}

	sb.WriteString("Please produce the final summary for your target audience.")
	return sb.String()
}

// Synthesize takes all analyst reports and produces the final
// audience-appropriate summary by calling the Anthropic API once more with
// a synthesizer persona and the aggregated findings.
func Synthesize(ctx context.Context, apiKey string, reports []AnalystReport, audience Audience) (string, error) {
	if _, ok := audienceConfigs[audience]; !ok {
		return "", fmt.Errorf("unknown audience %q", audience)
	}

	summary, err := callClaude(ctx, apiKey, buildSystemPrompt(audience), buildUserPrompt(reports), 1200)
	if err != nil {
		return "", fmt.Errorf("synthesizer: %w", err)
	}
	return summary, nil
}
