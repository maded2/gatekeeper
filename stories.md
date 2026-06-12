# Gatekeeper — User Stories

**Generated from:** `gatekeeper.md` (Product & Technical Specification v1.2.0), `quality_score_standard.md` (Engineering Standard v1.1.0)  
**Decomposition approach:** Elephant Carpaccio — thin vertical slices delivering end-to-end user value per story  
**Sequencing rationale:** Stories are ordered to reduce risk first (configuration and local feedback loop), then expand scope (differential analysis, CI gating, LLM enrichment), and finally harden edge cases (air-gapped mode, performance, error resilience).

---

## Personas

| Persona | Description | Primary Goal |
| --- | --- | --- |
| **Developer** | Individual engineer working locally on feature code | Get fast, actionable quality feedback before committing |
| **Code Reviewer / Dev Lead** | Senior engineer reviewing pull requests | Understand the quality impact of incoming changes at a glance |
| **CI/CD Operator** | Platform or DevOps engineer configuring pipelines | Obtain deterministic pass/fail signals to automate release gating |
| **Security Officer** | Organization security stakeholder | Ensure no secrets or sensitive data leave the organization's network |

---

## Epic A: Project Onboarding & Configuration

*These stories let a user install and configure Gatekeeper with zero confusion. Without these, no other story is possible.*

---

### Story A-1: Configure Gatekeeper for a New Project

**As a** Developer  
**I want** to initialize a `gatekeeper.json` configuration file in my project root  
**So that** Gatekeeper knows how to evaluate my codebase against my organization's quality standards

**Acceptance Criteria**:
- I can run a single command that creates a sensible default `gatekeeper.json` in my project root
- The default configuration uses safe, conservative thresholds (e.g., target score of 75)
- The default configuration includes common exclusion paths (`vendor/`, `node_modules/`, `dist/`)
- I can edit the file to customize thresholds, LLM provider settings, and exclusion paths
- If no `gatekeeper.json` exists when I run Gatekeeper, I receive a clear message explaining how to create one

**Definition of Done**:
- A new user can configure Gatekeeper in under 2 minutes with zero documentation beyond the CLI message
- The default configuration file is valid JSON and passes Gatekeeper's own validation

---

### Story A-2: Exclude Unwanted Paths from Analysis

**As a** Developer  
**I want** to exclude specific directories and file patterns from quality analysis  
**So that** generated code, third-party vendor libraries, and build artifacts do not inflate my quality score or waste analysis time

**Acceptance Criteria**:
- I can list path patterns (e.g., `**/vendor/**`, `**/generated/*.go`) in my configuration
- Files matching exclusion patterns are silently skipped during analysis
- I receive a warning if I exclude a path that contains no files (misconfigured pattern)
- Glob patterns support wildcard matching for nested directories

**Definition of Done**:
- Excluded paths are never scanned, parsed, or sent for evaluation
- A misconfigured exclusion pattern is surfaced as a warning, not an error

---

### Story A-3: Set a Custom Quality Score Threshold

**As a** CI/CD Operator  
**I want** to configure the minimum Quality Score threshold that triggers a pass or fail  
**So that** I can align Gatekeeper's gating behavior with my organization's release policy

**Acceptance Criteria**:
- I can set a numeric threshold (0–100) in my configuration
- Scores at or above the threshold result in a pass signal
- Scores below the threshold result in a fail signal
- The default threshold is 75 if I do not configure one
- I can enable a hard fail for any critical or high severity security finding, regardless of overall score

**Definition of Done**:
- Changing the threshold and re-running produces the expected pass/fail outcome
- The security hard-fail override works independently of the numeric threshold

---

## Epic B: Local Developer Feedback Loop

*These stories deliver the core value proposition: fast, actionable quality feedback for the developer sitting at their terminal.*

---

### Story B-1: Check Quality of My Entire Workspace

**As a** Developer  
**I want** to run a quality check on my entire working directory  
**So that** I can see my codebase's overall health before I commit or push my changes

**Acceptance Criteria**:
- I can run a single command to evaluate all source files in my current directory
- The output shows my overall Quality Score out of 100
- The output breaks down my score by pillar (Static Health, Architecture, Verification, Security)
- The output lists specific findings with file names and line numbers
- The command completes within 10 seconds for codebases up to 100,000 lines

**Definition of Done**:
- A developer can understand their codebase health in under 30 seconds from running the command to reading the output
- The pillar breakdown adds up to the total score

---

### Story B-2: See Readable Findings in My Terminal

**As a** Developer  
**I want** quality findings displayed as a formatted, color-coded table in my terminal  
**So that** I can quickly scan results and identify which issues need my attention

**Acceptance Criteria**:
- Findings are displayed in a tabular format with columns for priority, file location, and description
- High-priority findings are visually distinct (e.g., red highlighting) from low-priority findings
- Each finding includes the file path, line range, and a plain-language description of the issue
- Each finding includes an actionable recommendation for how to fix the issue
- The default output format is the human-readable table; I can opt into JSON output when needed

**Definition of Done**:
- A developer can identify their top 3 issues within 5 seconds of viewing the output
- Findings are ordered by priority, with the highest impact issues listed first

---

### Story B-3: Receive Actionable Remediation Suggestions

**As a** Developer  
**I want** each quality finding to include a specific, actionable fix  
**So that** I know exactly what to change without having to research the issue myself

**Acceptance Criteria**:
- Every finding includes a remediation suggestion written in plain language
- Remediations reference the specific code location (file and line range)
- Remediations are prioritized so I can tackle the highest-impact fixes first
- When a finding relates to naming conventions, the suggestion includes a recommended name
- When a finding relates to complexity, the suggestion describes what to extract or simplify

**Definition of Done**:
- A developer can apply a remediation without needing to consult external documentation
- Remediations are specific to the code found, not generic advice

---

### Story B-4: Skip Analysis for Trivial Changes

**As a** Developer  
**I want** Gatekeeper to skip analysis when my changes are purely documentation, whitespace, or non-code files  
**So that** I get instant feedback (sub-300ms) instead of waiting for an unnecessary full evaluation

**Acceptance Criteria**:
- Changes that affect only `.md`, `.txt`, or similar non-code files trigger an immediate pass
- Changes that are purely whitespace-only trigger an immediate pass
- I see a message explaining that the changes were classified as trivial and skipped
- The skip behavior does not apply when I explicitly request a full workspace check

**Definition of Done**:
- A commit containing only README changes returns a result in under 300 milliseconds
- The skip message is clear and does not confuse the developer

---

## Epic C: Differential Analysis (PR & Commit Workflows)

*These stories enable the Code Reviewer persona to assess the quality impact of incoming changes.*

---

### Story C-1: Check Quality of Changes Between Two Branches

**As a** Code Reviewer  
**I want** to evaluate the quality impact of the differences between a feature branch and its base branch  
**So that** I can assess whether a pull request introduces quality regressions before merging

**Acceptance Criteria**:
- I can specify a base branch and a target branch to compare
- Only the files and lines that differ between the two branches are analyzed
- The output shows how the changes affect each quality pillar
- If the changes introduce new complexity, the output highlights the specific functions affected
- If the changes improve quality, the output acknowledges the positive impact

**Definition of Done**:
- A code reviewer can determine whether a PR improves, degrades, or maintains quality at a glance
- The analysis focuses only on changed code, not the entire codebase

---

### Story C-2: Check Quality of a Specific Commit Range

**As a** Code Reviewer  
**I want** to evaluate the quality of a specific range of commits  
**So that** I can audit a batch of changes without analyzing unrelated work on the same branch

**Acceptance Criteria**:
- I can specify a commit range (e.g., `HEAD~3..HEAD`) to evaluate
- Only the changes introduced in the specified commits are analyzed
- The output aggregates findings across all commits in the range
- I can see which commit introduced each finding

**Definition of Done**:
- A reviewer can isolate and evaluate exactly the commits they care about
- Findings are attributable to specific commits within the range

---

### Story C-3: Generate a Markdown Summary for Pull Requests

**As a** Code Reviewer  
**I want** the quality check results formatted as a markdown summary  
**So that** I can paste the results directly into a pull request comment for team visibility

**Acceptance Criteria**:
- I can request markdown-formatted output
- The markdown includes the overall score, pillar breakdown, and top findings
- The markdown is formatted for readability in GitHub, GitLab, or similar platforms
- The output is concise enough to fit in a single PR comment without truncation

**Definition of Done**:
- The markdown output pastes cleanly into a PR comment with no formatting issues
- A reviewer can share quality results with their team in one copy-paste action

---

## Epic D: CI/CD Gating & Pipeline Integration

*These stories enable the CI/CD Operator to use Gatekeeper as an automated release gate.*

---

### Story D-1: Get a Deterministic Pass/Fail Signal for Pipelines

**As a** CI/CD Operator  
**I want** Gatekeeper to return a standard exit code that indicates pass, fail, or error  
**So that** my CI/CD pipeline can automatically decide whether to allow or block a deployment

**Acceptance Criteria**:
- Exit code 0 means the Quality Score meets or exceeds the configured threshold (pass)
- Exit code 2 means the Quality Score is below the threshold (quality gate blocked)
- Exit code 1 means an unexpected runtime error occurred (configuration missing, path unreadable)
- The exit code behavior is consistent and deterministic across runs
- The exit code is returned after all output is flushed to the terminal

**Definition of Done**:
- A CI/CD pipeline can gate a deployment solely on Gatekeeper's exit code
- The three exit codes cover all possible outcomes (pass, fail, error)

---

### Story D-2: Output Machine-Readable Results for Pipeline Logging

**As a** CI/CD Operator  
**I want** Gatekeeper to output results in structured JSON format  
**So that** my pipeline can log, store, and trend quality scores over time

**Acceptance Criteria**:
- I can request JSON output via a command-line flag
- The JSON includes the overall score, per-pillar scores, and all findings with locations
- The JSON structure is stable and documented so I can parse it reliably
- I can write the JSON output to a file for archival
- The JSON output includes a timestamp for correlation with pipeline runs

**Definition of Done**:
- A pipeline can parse the JSON output without custom scripting
- Historical quality scores can be reconstructed from archived JSON reports

---

### Story D-3: Block Deployment When Critical Security Issues Exist

**As a** Security Officer  
**I want** Gatekeeper to block deployment immediately when critical or high severity security vulnerabilities are found  
**So that** no code with known critical security flaws reaches production, regardless of overall score

**Acceptance Criteria**:
- I can enable a hard-fail toggle for critical security findings in my configuration
- When enabled, any critical or high severity security finding triggers a fail regardless of the overall Quality Score
- The output clearly identifies which security finding triggered the block
- This behavior is independent of the numeric quality threshold

**Definition of Done**:
- A codebase with a perfect quality score but one critical CVE is still blocked
- The security block reason is clearly communicated in the output

---

### Story D-4: Enforce Mutation Testing Coverage Requirements

**As a** CI/CD Operator  
**I want** the Verification pillar score to be penalized when mutation testing coverage falls below my organization's floor  
**So that** teams cannot pass the quality gate with hollow unit tests that lack meaningful assertions

**Acceptance Criteria**:
- I can configure a minimum mutation coverage threshold (default 80%)
- When mutation coverage falls below the threshold, the Verification pillar score is reduced by 20%
- The output explains that the penalty was applied due to insufficient mutation coverage
- When mutation coverage meets or exceeds the threshold, no penalty is applied

**Definition of Done**:
- A team with 70% mutation coverage sees a visibly lower Verification score than a team with 90%
- The penalty is clearly attributed to mutation coverage in the output

---

## Epic E: LLM-Enhanced Code Evaluation

*These stories add intelligent, context-aware analysis that goes beyond static rules.*

---

### Story E-1: Evaluate Code Readability and Intent

**As a** Code Reviewer  
**I want** Gatekeeper to evaluate whether variable names, function names, and code structure clearly express their intent  
**So that** I can catch confusing or cryptic naming before it becomes technical debt

**Acceptance Criteria**:
- Gatekeeper flags variables and functions with names that obscure their purpose (e.g., `a`, `temp`, `data2`)
- Each flag includes a suggested improvement that reflects the code's actual purpose
- The evaluation considers the surrounding context to understand what the code is doing
- Findings are specific to the code, not generic naming advice

**Definition of Done**:
- A developer receiving a readability flag understands both the problem and the fix
- The suggestions are contextually appropriate, not templated

---

### Story E-2: Detect Code Duplication and DRY Violations

**As a** Developer  
**I want** Gatekeeper to identify duplicated logic across my codebase  
**So that** I can extract shared behavior into reusable abstractions and reduce maintenance burden

**Acceptance Criteria**:
- Gatekeeper identifies blocks of logically identical or near-identical code across files
- Each finding shows both locations where the duplication exists
- The finding includes a recommendation for how to extract the shared logic
- Minor variations (e.g., different variable names but same structure) are still detected

**Definition of Done**:
- A developer can act on a duplication finding to consolidate the duplicated code
- False positives (e.g., boilerplate that is intentionally repeated) are minimized

---

### Story E-3: Configure a Custom LLM Provider for On-Premise Evaluation

**As a** Security Officer  
**I want** to point Gatekeeper at my organization's internal LLM endpoint  
**So that** no source code is transmitted to external cloud providers

**Acceptance Criteria**:
- I can configure a custom base URL for the LLM provider in my configuration
- I can specify which environment variable holds the API key
- I can set a timeout for LLM requests to prevent hanging
- I can configure the model name to match my internal deployment
- I can set the temperature to 0 for deterministic evaluation results

**Definition of Done**:
- Gatekeeper works with any OpenAI-compatible endpoint, including on-premise deployments like Ollama or vLLM
- No code is sent to any external service when an internal endpoint is configured

---

## Epic F: Privacy & Air-Gapped Operation

*These stories address the Security Officer's need to operate in restricted environments.*

---

### Story F-1: Prevent Secrets from Leaving My Network

**As a** Security Officer  
**I want** Gatekeeper to automatically detect and redact secrets, API keys, and credentials from code before any network transmission  
**So that** sensitive data is never exposed to the LLM provider, even accidentally

**Acceptance Criteria**:
- Gatekeeper scans code for patterns matching common secrets (API keys, passwords, tokens, certificates)
- Detected secrets are replaced with `[REDACTED]` placeholders before any network transmission
- The redaction happens locally and never sends the original secret value
- I can enable or disable this feature in my configuration
- I receive a warning if secrets are detected, even after redaction

**Definition of Done**:
- A developer accidentally committing an API key does not expose it to the LLM provider
- The developer is warned that secrets were found and redacted

---

### Story F-2: Disable Cloud Transmission Entirely

**As a** Security Officer  
**I want** to configure Gatekeeper to never transmit data to public cloud endpoints  
**So that** my organization's air-gapped policy is enforced at the tool level

**Acceptance Criteria**:
- I can set a configuration flag that prohibits all public cloud transmission
- When enabled, Gatekeeper refuses to connect to any public LLM endpoint
- When enabled and no internal LLM is configured, Gatekeeper falls back to rule-based evaluation only
- The fallback mode still computes a Quality Score using static analysis and metrics
- I receive a clear message when LLM-enhanced evaluation is unavailable and rule-based mode is active

**Definition of Done**:
- An air-gapped environment can run Gatekeeper and receive a Quality Score without any outbound network calls
- The user understands which evaluation capabilities are available in rule-only mode

---

### Story F-3: Transmit Only Changed Code, Not the Entire Codebase

**As a** Security Officer  
**I want** Gatekeeper to send only the specific code segments being evaluated, not surrounding project files  
**So that** the minimum necessary code leaves my environment for LLM analysis

**Acceptance Criteria**:
- When evaluating a diff or commit range, only the changed functions or classes are sent for LLM analysis
- Surrounding context is limited to the immediate parent scope (e.g., the containing function or class)
- Unchanged files are never transmitted
- I can see in the output which code segments were sent for evaluation

**Definition of Done**:
- A developer changing one function does not have their entire file transmitted
- The transmitted code is the minimum required for meaningful evaluation

---

## Epic G: Performance & Reliability

*These stories ensure Gatekeeper is fast enough for local use and resilient enough for pipeline use.*

---

### Story G-1: Get Fast Feedback on Small Changes

**As a** Developer  
**I want** quality checks on small commits (fewer than 10 files) to complete in under 300 milliseconds  
**So that** Gatekeeper feels instant and does not interrupt my development flow

**Acceptance Criteria**:
- Evaluating a commit with fewer than 10 changed files completes within 300 milliseconds
- The fast path applies when changes are simple and do not require LLM analysis
- I see results immediately without perceptible delay
- The speed guarantee applies to diff and commit-range modes

**Definition of Done**:
- A developer cannot perceive a delay when running Gatekeeper on a small commit
- The tool feels as fast as a linter

---

### Story G-2: Handle LLM Failures Gracefully

**As a** CI/CD Operator  
**I want** Gatekeeper to retry failed LLM requests and fall back to rule-based scoring if the LLM is unavailable  
**So that** my pipeline does not fail because of a transient network issue or LLM provider outage

**Acceptance Criteria**:
- If an LLM request fails, Gatekeeper retries up to 2 additional times
- If all retries fail, Gatekeeper computes the Quality Score using only static analysis and metrics
- The output clearly indicates when LLM-enhanced evaluation was unavailable
- The fallback score is computed without the pillars that require LLM input
- The pipeline receives a valid exit code even in fallback mode

**Definition of Done**:
- A transient LLM timeout does not cause the pipeline to fail
- The user understands which parts of the evaluation were completed and which were skipped

---

### Story G-3: Understand Clear Error Messages

**As a** Developer  
**I want** Gatekeeper to provide clear, actionable error messages when something goes wrong  
**So that** I can fix the problem quickly without guessing what went wrong

**Acceptance Criteria**:
- Missing configuration file: message explains where to place `gatekeeper.json`
- Invalid JSON in configuration: message identifies the line and nature of the error
- Unreachable LLM endpoint: message explains the connection failure and suggests checking the URL
- Missing API key environment variable: message names the expected variable
- Unreadable path: message identifies which path could not be accessed
- All error messages use exit code 1 (runtime error), distinguishing them from quality gate failures (exit code 2)

**Definition of Done**:
- A developer can resolve any error within one attempt using the error message alone
- Error messages never expose stack traces or internal implementation details

---

## Summary

| Epic | Focus | Stories | Primary Persona |
| --- | --- | --- | --- |
| **A: Project Onboarding & Configuration** | Setup, config, exclusions, thresholds | A-1 to A-3 | Developer, CI/CD Operator |
| **B: Local Developer Feedback Loop** | Workspace checks, terminal output, remediations, trivial skip | B-1 to B-4 | Developer |
| **C: Differential Analysis** | Branch diffs, commit ranges, PR markdown | C-1 to C-3 | Code Reviewer / Dev Lead |
| **D: CI/CD Gating** | Exit codes, JSON output, security blocking, mutation guardrail | D-1 to D-4 | CI/CD Operator, Security Officer |
| **E: LLM-Enhanced Evaluation** | Readability, duplication detection, custom provider | E-1 to E-3 | Code Reviewer, Developer, Security Officer |
| **F: Privacy & Air-Gapped** | Secret redaction, cloud prohibition, minimal transmission | F-1 to F-3 | Security Officer |
| **G: Performance & Reliability** | Speed, retry/fallback, error messages | G-1 to G-3 | Developer, CI/CD Operator |

**Total: 20 user stories** across 7 epics, 4 personas.

### How These Stories Solve the Original Problem

The original specification describes Gatekeeper as a quality scoring tool serving three operational modes (local CLI, PR review, CI gating). These 20 stories collectively cover:

1. **Onboarding (A):** A new user can install, configure, and run Gatekeeper in minutes
2. **Core value (B):** A developer gets fast, readable, actionable quality feedback locally
3. **Collaboration (C):** A reviewer assesses the quality impact of incoming changes
4. **Automation (D):** A pipeline gates releases on deterministic quality signals
5. **Intelligence (E):** LLM analysis catches issues that static rules cannot (readability, duplication)
6. **Security (F):** An organization operates Gatekeeper in air-gapped or highly restricted environments
7. **Resilience (G):** The tool is fast enough for local use and reliable enough for production pipelines

Each story is a thin vertical slice: configuring a threshold (A-3) delivers real value (pipelines can gate) without needing LLM integration, differential analysis, or secret redaction. Stories build incrementally toward the complete system described in the specification.
