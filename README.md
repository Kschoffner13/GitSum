# GitSum

GitSum is a CLI tool that analyses your git commit history using AI and produces summaries tailored to different audiences — engineers, managers, clients, and changelogs.

## Features

- **Multi-agent pipeline** — four specialist AI agents analyse your commits in parallel from different perspectives (technical, impact, risk, velocity), then a synthesizer agent combines their findings into a single coherent output
- **Audience-aware output** — the same commit history produces very different documents depending on who needs to read it
- **Fast** — analyst agents run concurrently, so the wall-clock time is roughly that of a single API call
- **Partial failure tolerance** — if one analyst fails (e.g. a rate limit), the pipeline continues with the remaining results

## Installation

```bash
git clone https://github.com/Kschoffner13/GitSum
cd GitSum
go build -o gitsum .
```

Set your Anthropic API key (get one at [console.anthropic.com/settings/keys](https://console.anthropic.com/settings/keys)).

Either store it once via gitsum itself — saved to a user-level config file (e.g. `%AppData%\gitsum\.env` on Windows), so it works from any directory and any shell session:

```bash
gitsum config set-key sk-ant-...

# Or omit the argument to be prompted instead, so the key never touches your shell history:
gitsum config set-key
```

…or export it manually for the current session (this takes precedence over the stored value):

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

## Commands

### `gitsum generate`

Produces a plain list of recent commits. No AI calls; useful as a quick sanity check.

```bash
gitsum generate
```

### `gitsum summary`

Runs the full multi-agent pipeline and produces an audience-appropriate summary.

```bash
# Default: lead-engineer audience, last 20 commits
gitsum summary

# Choose an audience
gitsum summary --audience client
gitsum summary --audience manager
gitsum summary --audience release-notes

# Limit to commits since a tag or ref
gitsum summary --audience release-notes --since v1.2.0
gitsum summary --audience manager --since HEAD~50

# Increase the commit window (default 20)
gitsum summary --limit 50

# Print raw analyst reports before the final summary
gitsum summary --verbose
```

#### Audiences

| Flag value | Best for | Emphasis |
|---|---|---|
| `lead-engineer` (default) | Tech leads, senior engineers | Technical detail + risk |
| `manager` | Engineering managers | Outcomes + team health |
| `client` | Non-technical stakeholders | Plain-language benefits, zero jargon |
| `release-notes` | Changelogs, public releases | Structured Added/Changed/Fixed/… |

### `gitsum version`

```bash
gitsum version
```

## How the pipeline works

```
git log → []Commit
              │
    ┌─────────┴──────────┐  (4 goroutines, parallel)
    │ Technical Analyst  │──┐
    │ Impact Analyst     │──┤
    │ Risk Analyst       │──┼──► []AnalystReport
    │ Velocity Analyst   │──┘
    └────────────────────┘
              │
    Synthesizer (audience-aware)
              │
         Final summary → stdout
```

Progress messages are written to **stderr** so that stdout remains clean for piping and redirection:

```bash
gitsum summary --audience release-notes --since v1.2.0 > CHANGELOG.md
```

## Project structure

```
.
├── main.go
├── cmd/
│   ├── root.go        # Root Cobra command
│   ├── gitsum.go      # version subcommand
│   ├── generate.go    # generate subcommand (plain commit list)
│   ├── summary.go     # summary subcommand (AI pipeline)
│   └── config.go      # config subcommand (set-key)
└── internal/
    ├── agents/
    │   ├── types.go        # Shared types: Commit, Audience, AnalystReport, SummaryResult
    │   ├── analyst.go      # Parallel analyst agents + Anthropic API client
    │   ├── synthesizer.go  # Audience-aware synthesizer agent
    │   └── pipeline.go     # Orchestrator: Run(), CommitsFromGit()
    ├── ai/
    │   └── summarize.go    # Legacy plain-text summarizer (used by generate)
    ├── config/
    │   └── config.go       # User-level API key storage (~/.config/gitsum or %AppData%\gitsum)
    ├── git/
    │   └── parser.go       # git log parsing → []git.Commit
    ├── output/
    │   └── writer.go       # io.Writer helper
    ├── tui/
    │   └── spinner.go      # (placeholder for future Bubbletea UI)
    └── version/
        └── version.go      # Build-time version string
```

## Extending

### Adding a new analyst perspective

1. Add a new `AnalystRole` constant in `internal/agents/types.go`
2. Add a new `analyst` entry in the `analysts` slice in `internal/agents/analyst.go` with a focused system prompt
3. Add the role label to the `roleLabels` map in `internal/agents/synthesizer.go`

### Adding a new audience

1. Add a new `Audience` constant in `internal/agents/types.go` and append it to `ValidAudiences`
2. Add an `audienceConfig` entry in `internal/agents/synthesizer.go`

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `ANTHROPIC_API_KEY` | Yes (for `summary`) | Your Anthropic API key |

## Building for release

The version string is embedded at build time:

```bash
go build -ldflags "-X github.com/Kschoffner13/GitSum/internal/version.Version=1.0.0" -o gitsum .
```
