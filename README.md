# Gatekeeper — Automated Quality Score Tool

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/doc/devel/release.html)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Gatekeeper** is a single-binary CLI tool that evaluates git commits, pull requests, or full codebases against an organizational Quality Score Standard. It combines static analysis, structural code metrics, and LLM-driven reasoning to produce a score out of 100 with actionable remediations.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Commands](#commands)
- [Configuration](#configuration)
- [Quality Score](#quality-score)
- [Architecture](#architecture)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Three Operational Modes**
  - Local CLI: `gatekeeper check` — fast feedback for developers
  - Differential: `gatekeeper diff --base=main --target=feature` — PR impact assessment
  - CI/CD Gating: deterministic exit codes (0 = pass, 1 = error, 2 = fail)

- **Four Quality Pillars**
  - Static Code Health (20 pts): formatting, linting, dead code
  - Engineering Architecture (25 pts): cyclomatic complexity, cognitive depth
  - Dynamic Verification (35 pts): Farley Index, mutation testing
  - Security & Supply Chain (20 pts): SAST, SCA, secret detection

- **Privacy-First Design**
  - Secret redaction before any network transmission
  - Air-gapped mode with rule-based fallback
  - Configurable LLM provider (any OpenAI-compatible endpoint)

- **Multiple Output Formats**
  - Pretty terminal tables (default)
  - JSON for pipeline logging
  - Markdown for PR comments

## Installation

### From Source

```bash
git clone https://github.com/yourorg/gatekeeper.git
cd gatekeeper
go build -o gatekeeper .
```

### Binary Release

Download the latest release from the [Releases](https://github.com/yourorg/gatekeeper/releases) page and place the binary in your `PATH`.

## Quick Start

```bash
# 1. Initialize configuration
gatekeeper init

# 2. Check your workspace quality
gatekeeper check

# 3. Evaluate changes between branches
gatekeeper diff --base=main --target=feature

# 4. Evaluate a commit range
gatekeeper commit-range --range=HEAD~3..HEAD
```

## Commands

### `gatekeeper init`

Creates a default `gatekeeper.json` configuration file in the current directory.

```bash
gatekeeper init
```

### `gatekeeper check [path]`

Evaluates all source files in the given path and outputs a quality score with findings.

```bash
gatekeeper check                          # current directory
gatekeeper check --path=./src             # specific path
gatekeeper check --format=json            # JSON output
gatekeeper check --format=markdown        # Markdown output
gatekeeper check --output=report.json     # Write to file (JSON)
```

### `gatekeeper diff`

Evaluates the quality impact of differences between a base and target branch.

```bash
gatekeeper diff --base=main --target=feature
gatekeeper diff --base=main --target=feature --format=json
```

### `gatekeeper commit-range`

Evaluates the quality of changes introduced in a specific range of commits.

```bash
gatekeeper commit-range --range=HEAD~3..HEAD
gatekeeper commit-range --range=abc123..def456 --format=markdown
```

## Configuration

Gatekeeper uses a `gatekeeper.json` file in the project root. Run `gatekeeper init` to create a default configuration.

```json
{
  "gatekeeper": {
    "target_threshold": 75.0,
    "fail_on_critical_security": true,
    "llm": {
      "provider": "openai-compatible",
      "base_url": "https://api.openai.com/v1",
      "model_name": "gpt-4o-mini",
      "api_key_env_var": "GATEKEEPER_API_KEY",
      "timeout_ms": 4000,
      "temperature": 0.0
    },
    "privacy": {
      "allow_public_cloud_transmission": false,
      "data_scrubbing": true
    }
  },
  "pillars": {
    "static": {
      "max_cyclomatic_complexity": 10,
      "allow_experimental_modules": false
    },
    "verification": {
      "require_mutation_testing": true,
      "mutation_coverage_floor": 80.0
    }
  },
  "exclusions": {
    "paths": [
      "**/vendor/**",
      "**/node_modules/**",
      "**/dist/**"
    ]
  }
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `target_threshold` | float | 75.0 | Minimum score to pass (0-100) |
| `fail_on_critical_security` | bool | true | Block on critical/high security findings |
| `llm.base_url` | string | - | LLM provider endpoint |
| `llm.model_name` | string | - | Model identifier |
| `llm.api_key_env_var` | string | - | Environment variable for API key |
| `llm.timeout_ms` | int | 4000 | Request timeout in milliseconds |
| `llm.temperature` | float | 0.0 | Model temperature (0 = deterministic) |
| `privacy.allow_public_cloud_transmission` | bool | false | Allow public cloud endpoints |
| `privacy.data_scrubbing` | bool | true | Redact secrets before transmission |
| `exclusions.paths` | []string | see above | Glob patterns to exclude |

## Quality Score

The Quality Score is calculated as a sum of four pillars, maximum 100 points:

| Pillar | Max Points | Description |
|--------|------------|-------------|
| Static Code Health | 20 | Formatting, linting, dead code elimination |
| Engineering Architecture | 25 | Cyclomatic complexity ≤ 10, cognitive depth ≤ 3 |
| Dynamic Verification | 35 | Farley Index × 3.5, with 0.8× mutation penalty |
| Security & Supply Chain | 20 | Zero critical/high SAST flaws, zero CVEs |

### Operational Tiers

| Score | Status | Pipeline Action |
|-------|--------|-----------------|
| 90–100 | Elite Tier | Unrestricted deployment |
| 75–89 | Target Standard | Approved for release |
| 60–74 | Technical Debt Alert | Warning, PRs that lower score blocked |
| 0–59 | Critical Remediate | Pipeline blocked, delivery halted |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Score ≥ threshold (pass) |
| 1 | Runtime error (config missing, invalid JSON, etc.) |
| 2 | Score < threshold (quality gate blocked) |

## Architecture

```
gatekeeper/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go             # Root command, slog init
│   ├── init.go             # Initialize configuration
│   ├── check.go            # Full workspace evaluation
│   ├── diff.go             # Branch diff evaluation
│   └── commit-range.go     # Commit range evaluation
├── internal/
│   ├── config/             # Configuration loading, validation
│   ├── evaluator/          # Quality scoring, pillar computation
│   ├── git/                # Git diff/range extraction
│   ├── llm/                # LLM config, secret scrubbing, retry logic
│   ├── reporter/           # Output formatters (pretty, JSON, markdown)
│   └── scanner/            # File discovery with glob exclusions
├── pkg/
│   └── score/              # Public score types and pillar definitions
└── tests/
    └── integration/        # End-to-end CLI tests
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests first (ATDD methodology)
4. Commit your changes (`git commit -m 'feat: add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

### Development

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run integration tests
go test ./tests/integration/ -v

# Build the binary
go build -o gatekeeper .

# Run the binary
./gatekeeper check
```

### Code Style

- Use `gofmt` for formatting
- Run `go vet ./...` before committing
- Follow the [Effective Go](https://go.dev/doc/effective_go) guidelines
- Write tests for all new functionality
- Use `log/slog` for structured logging

## License

[MIT License](LICENSE)
