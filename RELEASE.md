# Gatekeeper v0.1.0

First release of **Gatekeeper** — a single-binary CLI tool that evaluates git commits, pull requests, or full codebases against an organizational Quality Score Standard.

## Features

- **Three Operational Modes**
  - `gatekeeper check` — full workspace evaluation
  - `gatekeeper diff` — branch-to-branch comparison
  - `gatekeeper commit-range` — isolated commit range analysis

- **Four Quality Pillars** (max 100 points)
  - Static Code Health (20 pts): formatting, linting, dead code
  - Engineering Architecture (25 pts): cyclomatic complexity, cognitive depth
  - Dynamic Verification (35 pts): Farley Index, mutation testing
  - Security & Supply Chain (20 pts): SAST, SCA, secret detection

- **Flexible Authentication**
  - API key via environment variables
  - Browser-based OAuth2 login for Google Gemini

- **Privacy-First Design**
  - Secret redaction before network transmission
  - Air-gapped mode with rule-based fallback

- **Multiple Output Formats**
  - Pretty terminal tables (default)
  - JSON for CI/CD pipelines
  - Markdown for PR comments

## Quick Start

```bash
# Initialize configuration
gatekeeper init

# Check workspace quality
gatekeeper check

# Evaluate changes between branches
gatekeeper diff --base=main --target=feature
```

## Binary Downloads

| Platform | Binary |
|----------|--------|
| Linux (amd64) | `gatekeeper-linux-amd64` |
| Linux (arm64) | `gatekeeper-linux-arm64` |
| macOS (amd64) | `gatekeeper-darwin-amd64` |
| macOS (arm64) | `gatekeeper-darwin-arm64` |
| Windows (amd64) | `gatekeeper-windows-amd64.exe` |

All binaries are statically linked and stripped. Verify checksums with `SHA256SUMS`.

## Build from Source

```bash
git clone https://github.com/maded2/gatekeeper.git
cd gatekeeper
make build          # current platform
make release        # all platforms
```
