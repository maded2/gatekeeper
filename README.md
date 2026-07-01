# Gatekeeper — Automated Quality Score Tool

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/doc/devel/release.html)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

**Gatekeeper** is a single-binary CLI tool that evaluates git commits, pull requests, or full codebases against an organizational Quality Score Standard. It combines static analysis, structural code metrics, and LLM-driven reasoning to produce a score out of 100 with actionable remediations.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Build](#build)
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

- **Flexible Authentication**
  - API key authentication via environment variables
  - Browser-based OAuth2 login for Google Gemini and other providers
  - Token caching and automatic refresh

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
make build
```

This produces a statically linked binary in `dist/` for your current platform.

### Cross-Platform Builds

```bash
# Build for all platforms (Linux, macOS, Windows)
make release

# Build for a specific platform
make linux    # linux-amd64, linux-arm64
make macos    # darwin-amd64, darwin-arm64
make windows  # windows-amd64.exe

# Clean build artifacts
make clean
```

All binaries are statically linked, stripped, and embed version metadata.

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

### API Key Authentication (Default)

```json
{
  "gatekeeper": {
    "target_threshold": 75.0,
    "fail_on_critical_security": true,
    "llm": {
      "provider": "openai-compatible",
      "base_url": "https://api.openai.com/v1",
      "model_name": "gpt-4o-mini",
      "auth_type": "api_key",
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

### Browser OAuth Authentication (Google Gemini)

For providers that support OAuth2 (such as Google Gemini), use browser-based login. Gatekeeper opens your browser for authentication and caches the resulting token.

```json
{
  "gatekeeper": {
    "target_threshold": 75.0,
    "llm": {
      "base_url": "https://generativelanguage.googleapis.com/v1beta",
      "model_name": "gemini-pro",
      "auth_type": "oauth_browser",
      "oauth_token_url": "https://oauth2.googleapis.com/token",
      "oauth_auth_url": "https://accounts.google.com/o/oauth2/v2/auth",
      "oauth_client_id_env_var": "GOOGLE_CLIENT_ID",
      "oauth_client_secret_env_var": "GOOGLE_CLIENT_SECRET",
      "oauth_scopes": ["https://www.googleapis.com/auth/cloud-platform"],
      "oauth_redirect_url": "http://localhost:8080/callback",
      "oauth_token_cache_file": "~/.cache/gatekeeper/oauth_token.json",
      "timeout_ms": 5000,
      "temperature": 0.0
    }
  }
}
```

The OAuth flow works as follows:
1. Gatekeeper opens your default browser to the provider's consent page
2. You sign in and grant permissions in the browser
3. The provider redirects to a local callback server (`localhost:8080/callback`)
4. Gatekeeper exchanges the authorization code for an access token
5. The token is cached to disk and automatically refreshed when it expires

Styled HTML pages are served locally during the flow: a waiting page while you authenticate, a success confirmation after sign-in, and an error page if authentication fails.

**Token caching:** Tokens are persisted to the path specified by `oauth_token_cache_file`. On subsequent runs, Gatekeeper loads the cached token and refreshes it transparently if it has expired. If refresh fails (e.g., revoked consent), the browser flow is retried automatically.

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `target_threshold` | float | 75.0 | Minimum score to pass (0-100) |
| `fail_on_critical_security` | bool | true | Block on critical/high security findings |
| `llm.auth_type` | string | `api_key` | Authentication method: `api_key` or `oauth_browser` |
| `llm.base_url` | string | - | LLM provider endpoint |
| `llm.model_name` | string | - | Model identifier |
| `llm.api_key_env_var` | string | - | Environment variable for API key (api_key auth) |
| `llm.timeout_ms` | int | 4000 | Request timeout in milliseconds |
| `llm.temperature` | float | 0.0 | Model temperature (0 = deterministic) |
| `llm.oauth_token_url` | string | - | OAuth2 token endpoint (oauth_browser auth) |
| `llm.oauth_auth_url` | string | - | OAuth2 authorization endpoint (oauth_browser auth) |
| `llm.oauth_client_id_env_var` | string | - | Env var for OAuth client ID |
| `llm.oauth_client_secret_env_var` | string | - | Env var for OAuth client secret |
| `llm.oauth_scopes` | []string | - | OAuth2 scopes to request |
| `llm.oauth_redirect_url` | string | - | Local callback URL (e.g., `http://localhost:8080/callback`) |
| `llm.oauth_token_cache_file` | string | - | Path to cache OAuth tokens (supports `~` expansion) |
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
make test

# Run with verbose output
go test ./... -v

# Run integration tests
go test ./tests/integration/ -v

# Build for current platform
make build

# Build for all platforms
make release

# Run the binary
./dist/gatekeeper-linux-amd64 check
```

### Code Style

- Use `gofmt` for formatting
- Run `go vet ./...` before committing
- Follow the [Effective Go](https://go.dev/doc/effective_go) guidelines
- Write tests for all new functionality
- Use `log/slog` for structured logging

## License

[Apache License 2.0](LICENSE)
