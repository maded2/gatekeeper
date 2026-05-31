# 🛡️ Gatekeeper

**An LLM-powered automated code quality gate** that evaluates code diffs and enforces a quality threshold — locally (advisory) or in CI (authoritative).

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## The Problem

Software teams lack an automated, consistent, and objective gate that evaluates every pull request against meaningful quality criteria. Manual code reviews are inconsistent, fatigued, and often superficial. Linting tools catch syntax but not architectural soundness. CI catches too late and creates a push-and-pray dynamic.

## What Gatekeeper Is

Gatekeeper is an **LLM-based evaluation agent** that:

- Evaluates **code diffs** (not entire codebases) against quality criteria
- Produces a **Farley Score** (1–10) aggregating code quality, test coverage adequacy, and deployability
- **Rejects** submissions below a configurable threshold in CI (authoritative mode)
- **Warns** with actionable feedback when run locally (advisory mode)
- Uses the **same criteria and scoring** in both environments — only the consequence differs

## What Gatekeeper Is NOT

- Not a replacement for human code reviewers (architectural/design review)
- Not a full-codebase evaluator (diffs only)
- Not a code generator or fixer
- Not a test coverage generator (consumes existing coverage data)
- Not a security scanner (SAST/DAST is separate)
- Not a performance profiler

## Core Evaluation Pillars

1. **Code Quality** — structural problems, anti-patterns, maintainability
2. **Test Coverage Adequacy** — whether changed code paths are tested (not just numeric coverage)
3. **Deployability** — readiness for production deployment

## Quick Start

### Prerequisites

- Go 1.25+
- Git
- LLM API key (e.g., OpenAI) or self-hosted model endpoint

### Installation

```bash
go install github.com/eddie/gatekeeper/cmd/gatekeeper@latest
```

### Configuration

Create a `.gatekeeper.yml` file in your repository root:

```yaml
# Gatekeeper quality gate configuration
preset: balanced  # strict | balanced | permissive
threshold: 6.0

exclude:
  - "*.lock"
  - "vendor/*"
  - "node_modules/*"

pillar_weights:
  code_quality: 0.4
  test_coverage: 0.35
  deployability: 0.25
```

### Usage

**Local (advisory mode)** — check your uncommitted changes:

```bash
export GATEKEEPER_API_KEY="your-api-key"
gatekeeper check
```

**Local with JSON output** (for IDE integration):

```bash
gatekeeper check --format json
```

**CI (authoritative mode)** — evaluate a PR diff:

```bash
gatekeeper check --base main --format ci
```

**Evaluate a diff file**:

```bash
git diff HEAD...main | gatekeeper check --diff-file -
```

## Farley Score

The Farley Score is a single metric (1–10) that aggregates all three evaluation pillars:

| Score | Meaning |
|-------|---------|
| 9–10 | Exemplary quality across all pillars |
| 7–8 | Good quality, minor improvements possible |
| 6 | Meets minimum threshold |
| 4–5 | Below threshold — improvements needed |
| 1–3 | Critical failures across multiple pillars |

### Score Breakdown

```
🛡️  Gatekeeper Quality Check

✅ Farley Score: 7.5 / 10.0 (threshold: 6.0)

Pillar Breakdown:
  📊 Code Quality:    8.0 (weight: 40%)
  🧪 Test Coverage:   7.0 (weight: 35%)
  🚀 Deployability:   7.5 (weight: 25%)

✅ Code quality meets threshold — safe to push
```

## Configuration Presets

| Preset | Threshold | Description |
|--------|-----------|-------------|
| **strict** | 8.0 | Maximum quality enforcement — for teams with mature CI/CD and experienced developers |
| **balanced** | 6.0 | Default balanced approach — suitable for most teams |
| **permissive** | 4.0 | Lightweight gatekeeping — for early-stage projects or teams new to automated quality gates |

## Two Modes

### Advisory Mode (Local)

- Runs on the developer's workstation
- Evaluates uncommitted, staged, or arbitrary diffs
- **Never blocks** — warns with specific improvement recommendations
- Near-instant feedback (seconds)
- Optional JSON output for IDE integration

### Authoritative Mode (CI)

- Runs in CI/CD pipelines on pull requests
- **Hard rejects** below threshold — blocks merge via non-zero exit code
- Posts detailed feedback as PR comments
- Produces CI step summaries
- Branch protection integration

## Reliability & Fallback

When the LLM API is unavailable, Gatekeeper activates a **fallback mode** that uses deterministic rules:

- Hardcoded credential detection
- Large file warnings
- Basic pattern matching

Fallback results are clearly labeled and treated as **advisory** (not authoritative) in CI mode.

## Security

- API credentials are configured through **environment variables only** — never written to disk
- Sensitive data (API keys, passwords, tokens) is **redacted** from diffs before sending to the LLM API
- **Self-hosted model** support — configure a local endpoint so code diffs never leave your infrastructure
- Evaluation tracking is **opt-in** and records only metadata (no code content)

## Project Structure

```
gatekeeper/
├── cmd/gatekeeper/           # CLI entry point
├── internal/
│   ├── config/               # Configuration loading, validation, presets
│   ├── diff/                 # Diff capture (working dir, staged, base ref, file)
│   ├── evaluator/            # Quality, coverage, deployability evaluation
│   ├── farleyscore/          # Farley Score computation
│   ├── fallback/             # Deterministic fallback when LLM unavailable
│   ├── output/               # Text, JSON, CI summary, PR comments
│   ├── security/             # Credential management, sensitive data redaction
│   └── tracking/             # Evaluation recording, trend reports
├── go.mod
├── README.md
├── LICENSE
├── agent.md                  # Project context and implementation rules
├── problem-analysis.md       # Complete problem specification
└── user-stories.md           # 39 user stories across 11 epics
```

## Tech Stack

- **Primary language:** Golang
- **LLM client framework:** [CloudWeGo Eino](https://github.com/cloudwego/eino)
- **Runtime:** Go (cross-platform: macOS, Linux, CI containers)

## Key Constraints

| Constraint | Value |
|---|---|
| Local latency | < 10 seconds for typical diffs (< 500 lines) |
| CI latency | < 60 seconds for typical diffs |
| Default threshold | 6/10 (Farley Score) |
| Evaluation scope | Diff / change set only |
| Portability | Identical results on macOS, Linux, CI containers |

## Development

### Running Tests

```bash
go test ./...
```

### Adding a New Evaluation Check

1. Add the check function in `internal/evaluator/evaluator.go`
2. Add acceptance tests in `internal/evaluator/evaluator_test.go`
3. Run `go test ./...` to verify

## License

[MIT License](LICENSE)
