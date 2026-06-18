package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	anthropicAPI     = "https://api.anthropic.com/v1/messages"
	anthropicVersion = "2023-06-01"
	claudeModel      = "claude-sonnet-4-6"
)

// analyst pairs an [AnalystRole] with the system prompt that defines its
// analytical perspective.
type analyst struct {
	role         AnalystRole
	systemPrompt string
}

// analysts is the fixed set of parallel perspectives run against every commit
// history. Add a new entry here to introduce a new analytical lens.
var analysts = []analyst{
	{
		role: RoleTechnical,
		systemPrompt: `You are a senior engineer performing a technical code review of a git commit history.
Your job: identify WHAT changed at the code level.
Focus on:
- Which files, modules, or packages were modified
- Whether changes are new features, bug fixes, refactors, or dependency updates
- Architectural decisions or patterns introduced
- Any notable implementation details worth flagging
Be precise and terse. Use bullet points. No fluff. Output markdown.`,
	},
	{
		role: RoleImpact,
		systemPrompt: `You are a product analyst reviewing a git commit history.
Your job: identify WHAT THIS MEANS FUNCTIONALLY — bridge code changes to user-visible or system-visible behaviour.
Focus on:
- Which features or capabilities are added, changed, or removed
- How end-users or downstream consumers are affected
- API or interface changes
- Any behaviour that is now different from before
Write clearly without assuming deep technical knowledge. Output markdown.`,
	},
	{
		role: RoleRisk,
		systemPrompt: `You are a release engineer reviewing a git commit history for risk.
Your job: surface WHAT COULD GO WRONG or NEEDS ATTENTION before/after deployment.
Focus on:
- Breaking changes (API contracts, data schemas, config formats)
- Dependency version bumps and their implications
- Missing tests, error handling gaps, or TODOs left in code
- Migration steps required
- Anything that could affect stability, security, or rollback ability
Be specific. If something looks clean, say so briefly. Output markdown.`,
	},
	{
		role: RoleVelocity,
		systemPrompt: `You are an engineering manager reviewing a git commit history for team health signals.
Your job: characterise HOW the work was done, not what was done.
Focus on:
- Commit cadence and distribution over time
- Author contribution spread (if multiple authors)
- Commit message quality and consistency
- Signs of churn (repeated edits to same files), rushed work, or clean incremental progress
- Overall estimate: was this a focused sprint, scattered effort, or steady drumbeat?
Keep it factual. Output markdown.`,
	},
}

// ---- Anthropic API types ------------------------------------------------

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ---- Internal helpers ---------------------------------------------------

// callClaude sends a single prompt to the Anthropic API and returns the
// plain-text response. maxTokens caps the response length.
func callClaude(ctx context.Context, apiKey, systemPrompt, userPrompt string, maxTokens int) (string, error) {
	reqBody := anthropicRequest{
		Model:     claudeModel,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages:  []anthropicMessage{{Role: "user", Content: userPrompt}},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPI, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	var result anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("anthropic API error: %s", result.Error.Message)
	}

	var sb strings.Builder
	for _, block := range result.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	return sb.String(), nil
}

// formatCommits serialises the commit slice into a readable block suitable
// for inclusion in a prompt.
func formatCommits(commits []Commit) string {
	var sb strings.Builder
	for i, c := range commits {
		fmt.Fprintf(&sb, "--- Commit %d ---\n", i+1)
		fmt.Fprintf(&sb, "Hash:      %s\n", c.Hash)
		fmt.Fprintf(&sb, "Author:    %s\n", c.Author)
		fmt.Fprintf(&sb, "Timestamp: %s\n", c.Timestamp)
		fmt.Fprintf(&sb, "Message:   %s\n", c.Message)
		if c.Diff != "" {
			sb.WriteString("Diff (stat):\n")
			sb.WriteString(c.Diff)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// ---- Public API ---------------------------------------------------------

// RunAnalysts fans out all analyst agents in parallel and collects their
// reports. It never cancels on a single failure — partial errors are captured
// inside each [AnalystReport].Err so the synthesizer can still run with
// whatever results are available.
func RunAnalysts(ctx context.Context, apiKey string, commits []Commit) []AnalystReport {
	userPrompt := "Here is the git commit history to analyse:\n\n" + formatCommits(commits)

	results := make([]AnalystReport, len(analysts))
	done := make(chan struct{}, len(analysts))

	for i, a := range analysts {
		i, a := i, a // capture loop variables for the goroutine
		go func() {
			defer func() { done <- struct{}{} }()
			findings, err := callClaude(ctx, apiKey, a.systemPrompt, userPrompt, 800)
			results[i] = AnalystReport{
				Role:     a.role,
				Findings: findings,
				Err:      err,
			}
		}()
	}

	for range analysts {
		<-done
	}

	return results
}
