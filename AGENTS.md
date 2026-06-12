# Gatekeeper — Agent Context

This file is read by the coding agent before every session. It defines project conventions, stack decisions, and structural expectations.

---

## Project Summary

**Gatekeeper** is a single-binary CLI tool that evaluates git commits, pull requests, or full codebases against an organizational Quality Score Standard. It combines static analysis, structural code metrics, and LLM-driven reasoning to produce a score out of 100 with actionable remediations.

**Specs:**
- `gatekeeper.md` — Product & Technical Specification
- `quality_score_standard.md` — Quality Score algorithm and pillar definitions
- `stores.md` — Prioritized user stories

**Three operational modes:**
1. **Local CLI** — `gatekeeper check` — fast feedback for developers
2. **Differential** — `gatekeeper diff --base=main --target=feature` — PR impact assessment
3. **CI/CD Gating** — deterministic exit codes (0 = pass, 1 = error, 2 = fail)

---

## Stack & Conventions

### Language: Go (1.21+)

- Use the **standard library** wherever possible before reaching for third-party packages
- Module path: `gatekeeper` (update once `go.mod` is initialized)
- All code must compile with `go vet` and `gofmt` clean
- Prefer `err != nil` checks over panic; never swallow errors silently

### Logging: `log/slog` Only

- Use Go's built-in `log/slog` for **all** logging — no third-party logging libraries
- Logger is initialized once in `cmd/root.go` and passed down via context or struct fields
- Use structured attributes, not string formatting:

```go
slog.Info("evaluating diff",
    slog.String("base", baseBranch),
    slog.String("target", targetBranch),
    slog.Int("files", len(files)),
)

slog.Error("failed to parse config",
    slog.Path("path", configPath),
    slog.Any("err", err),
)
```

- Log levels:
  - `Debug` — only when `--verbose` flag is set
  - `Info` — normal operational events (started evaluation, completed, skipped trivial)
  - `Warn` — recoverable issues (config defaults applied, LLM fallback activated)
  - `Error` — failures that affect the current operation but don't crash the binary
- Never log secrets, API keys, or raw source code to logs
- Use `slog.With()` to attach persistent context (e.g., request ID, config path) to child loggers

### LLM Framework: Eino

- Use the **Eino** framework (`github.com/cloudwego/eino`) for all LLM orchestration
- Eino handles: ChatModel instantiation, prompt templating, structured JSON output parsing, and retry chains
- LLM provider is configured via `gatekeeper.json` and targets any **OpenAI-compatible** endpoint
- Key Eino components:
  - `ChatModel` — instantiated from config (`base_url`, `model_name`, `api_key_env_var`, `temperature`, `timeout_ms`)
  - `PromptTemplate` — injects AST code chunks and quality rules as system context
  - `StructuredOutput` — enforces JSON response shape matching the `LLMEvaluation` Go struct
  - `RetryMiddleware` — 2-retry fallback on LLM failures, then falls back to rule-based scoring
- LLM requests must respect:
  - **Air-gapped mode** — when `privacy.allow_public_cloud_transmission` is `false`, only internal endpoints are used
  - **Secret scrubbing** — code chunks are regex-redacted before leaving the local process
  - **Minimal transmission** — only changed functions/classes are sent, never full files

---

## Project Structure (Target)

```
gatekeeper/
  cmd/
    root.go          # slog init, cobra root command, Viper config loading
    check.go         # "gatekeeper check" — full workspace evaluation
    diff.go          # "gatekeeper diff" — branch-to-branch comparison
    commit-range.go  # "gatekeeper commit-range" — isolated commit range
  internal/
    config/          # gatekeeper.json parsing, validation, GatekeeperConfig struct
    scanner/         # file discovery, glob exclusions, .gitignore respect
    git/             # go-git diff extraction, commit range resolution
    ast/             # tree-sitter parent scope lookup, complexity metrics
    evaluator/       # pillar score computation, QualityCalculator, exit code mapping
    llm/             # Eino ChatModel factory, prompt templates, structured output, scrubbing
    reporter/        # Reporter interface, TerminalReporter (lipgloss), JSONReporter
  pkg/
    score/           # public Quality Score types and pillar definitions (if needed externally
```

### Package Naming

- Use `internal/` for all private packages — they must not be importable externally
- Use `pkg/` only for types or utilities that may be shared (e.g., score struct definitions)
- Keep packages small and cohesive; each package owns one concern

---

## Configuration: `gatekeeper.json`

Located at the project root. Schema defined in `gatekeeper.md` §6. Key sections:

- `gatekeeper.target_threshold` — pass/fail threshold (default 75)
- `gatekeeper.fail_on_critical_security` — hard-fail on critical CVEs
- `gatekeeper.llm` — provider URL, model, API key env var, timeout, temperature
- `gatekeeper.privacy` — cloud transmission toggle, data scrubbing toggle
- `pillars.static` — cyclomatic complexity limits
- `pillars.verification` — mutation testing requirements
- `exclusions.paths` — glob patterns to skip

---

## Quality Score Algorithm

Defined in `quality_score_standard.md` §4. Four pillars, max 100:

| Pillar | Max Points | Source |
|---|---|---|
| Static Code Health | 20 | Formatters, linters, dead code |
| Engineering Architecture | 25 | Cyclomatic complexity ≤ 10, cognitive depth ≤ 3 |
| Dynamic Verification | 35 | Farley Index × 3.5, with 0.8× mutation penalty |
| Security & Supply Chain | 20 | SAST (0 critical/high), SCA (0 CVEs) |

Exit codes: `0` = score ≥ threshold, `1` = runtime error, `2` = score < threshold.

---

## LLM Structured Output Shape

The LLM must return JSON matching this structure (enforced by Eino structured output):

```json
{
  "pillar_adjustments": {
    "static_health_deduction": 1.5,
    "architecture_deduction": 0.0
  },
  "remediations": [
    {
      "priority": "HIGH",
      "pillar": "static",
      "location": "src/auth.go:lines 12-24",
      "finding": "...",
      "actionable_fix": "..."
    }
  ]
}
```

---

## Key Third-Party Dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI command scaffolding |
| `github.com/spf13/viper` | JSON config loading |
| `github.com/cloudwego/eino` | LLM orchestration, templating, structured output |
| `github.com/go-git/go-git/v5` | Git diff and commit range extraction |
| `github.com/smacker/go-tree-sitter` | AST parsing, parent scope lookup |
| `github.com/charmbracelet/lipgloss` | Terminal table styling |
| `github.com/bmatcuk/doublestar/v4` | Glob pattern matching for exclusions |

---

## Rules for the Agent

1. **Always use `slog`** — never `fmt.Println` for operational output, never import `log` (pre-slog), never use third-party loggers
2. **Always use Eino** for LLM interactions — never construct raw HTTP calls to LLM endpoints
3. **Pass errors explicitly** — no global error handling, no `panic()` in library code
4. **Keep the binary fast** — diff checks on < 10 files must complete in < 300ms
5. **Never transmit secrets** — scrub code before LLM calls, respect air-gapped mode
6. **Write tests** — every new package gets unit tests; use `testify` or stdlib `testing`
7. **Read the specs first** — `gatekeeper.md` and `quality_score_standard.md` are the source of truth for behavior
8. **Commit after every story** — when a user story from `stores.md` is completed, stage all changes and commit to git with a message referencing the story ID (e.g., `feat: implement Story B-1 — check quality of entire workspace`)
