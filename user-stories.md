# Gatekeeper — Epics & User Stories Plan

## Problem Summary

Software teams practicing continuous integration lack an automated, consistent, and objective gate that evaluates every pull request against meaningful quality criteria and enforces a hard rejection when quality falls below a defensible threshold. Manual code reviews are inconsistent, fatigued, and often superficial. Linting tools catch syntax but not architectural soundness. CI-only feedback creates a "push-and-pray" dynamic where developers discover quality issues only after pushing.

**Gatekeeper** is an LLM-powered automated code quality gate that evaluates code diffs (not entire codebases) against three pillars — code quality, test coverage adequacy, and deployability — produces a Farley Score (1–10), and either warns (local/advisory mode) or rejects (CI/authoritative mode) based on a configurable threshold.

---

## User Personas

| Persona | Description |
|---|---|
| **Developer** | Individual contributor writing code. Wants fast, fair, transparent feedback before pushing. Fears false rejections blocking their work. |
| **Code Reviewer / Senior Engineer** | Reviews PRs before merge. Wants to focus on high-value architectural review, not catching basic mistakes. Needs to validate or challenge automated results. |
| **Engineering Manager / Tech Lead** | Owns team quality standards. Needs consistent enforcement without adding headcount. Wants visibility into quality trends. |
| **DevOps / Platform Engineer** | Manages CI/CD pipeline reliability, latency, and cost. Concerned about flaky gates and API outages blocking merges. |
| **QA / Test Engineer** | Owns test quality standards. Wants assurance that test coverage analysis is meaningful, not just numeric. |
| **Security / Compliance Officer** | Concerned about code diffs sent to external APIs. Needs data handling and privacy guarantees. |
| **Release Manager** | Needs confidence that what's on main is deployable. Wants clear go/no-go signals. |

---

## Decomposition Approach

Stories are organized into **9 Epics** using the **Elephant Carpaccio** technique — each epic is sliced into the thinnest possible vertical slices that still deliver end-to-end user value. Stories within each epic are sequenced to:

1. **Reduce risk early** — tackle unknowns (LLM evaluation reliability, scoring consistency) before building on them
2. **Deliver value incrementally** — each story should be independently usable
3. **Build from foundation to polish** — configuration before customization, core evaluation before reporting

---

## Epic 1: Project Foundation & Configuration

**Goal:** Establish the shared configuration system that ensures local and CI evaluations use identical rules — the foundation for "same diff, same score."

**Rationale:** Without shared configuration, local and CI evaluations diverge immediately, destroying trust. This must be built first.

---

### Story 1.1: Repository Configuration File

**Story Title**: Repository Quality Gate Configuration

**As a** Developer or Engineering Manager
**I want** to define quality gate settings in a configuration file checked into my repository
**So that** both local and CI evaluations use identical rules without drift

**Acceptance Criteria**:
- A configuration file exists in the repository root that defines quality gate settings
- The configuration file supports specifying a quality threshold (numeric, 1–10 scale)
- The configuration file supports specifying file patterns to exclude from evaluation (e.g., lock files, generated code, binary assets)
- The configuration file supports specifying evaluation pillar weights (code quality, test coverage, deployability)
- The configuration file uses a human-readable format that developers can edit without special tools
- The configuration file has a default set of values when no file is present
- The tool reads the configuration file from the repository root before performing any evaluation

**Definition of Done**:
- Configuration file can be created with all supported settings
- Default configuration is applied when no file exists
- Configuration is loaded consistently from the same location in both local and CI contexts

---

### Story 1.2: Configuration Validation

**Story Title**: Configuration File Validation

**As a** Developer
**I want** to know immediately when my quality gate configuration is invalid
**So that** I can fix configuration errors before they cause silent misbehavior

**Acceptance Criteria**:
- The tool validates the configuration file format before any evaluation begins
- Invalid configuration produces a clear error message identifying the specific problem and its location
- Missing required fields produce a specific error indicating which field is missing
- Out-of-range values (e.g., threshold of 15) produce an error with the valid range
- The tool does not attempt evaluation when configuration is invalid
- A valid configuration file allows evaluation to proceed without warnings

**Definition of Done**:
- All configuration validation errors produce actionable error messages
- Valid configurations pass silently and allow evaluation to proceed
- Configuration errors are surfaced before any LLM calls or network requests

---

### Story 1.3: Configuration Presets

**Story Title**: Pre-built Configuration Presets

**As a** Engineering Manager
**I want** to start with pre-built configuration presets for common team profiles
**So that** I can set up the quality gate quickly without researching optimal settings

**Acceptance Criteria**:
- At least three pre-built presets are available (e.g., "strict," "balanced," "permissive")
- Each preset has a clearly documented description of its intended use case and trade-offs
- A preset can be referenced in the configuration file to apply its settings
- Individual settings can be overridden after applying a preset
- The default behavior uses a "balanced" preset when no preset is explicitly chosen
- Preset descriptions are visible when the tool is invoked with a help or list command

**Definition of Done**:
- All presets are documented with use cases and trade-offs
- Presets can be applied and overridden in the configuration file
- Help output lists available presets with descriptions

---

## Epic 2: Core Diff Evaluation

**Goal:** Build the ability to evaluate code diffs for quality — the heart of Gatekeeper.

**Rationale:** This is the core capability. Without it, nothing else matters. Built incrementally: first diff ingestion, then quality evaluation, then coverage assessment, then deployability.

---

### Story 2.1: Diff Input — Working Directory Changes

**Story Title**: Evaluate Uncommitted Working Directory Changes

**As a** Developer
**I want** to run the quality gate against my uncommitted changes in the working directory
**So that** I can self-check code quality at any point during development without committing

**Acceptance Criteria**:
- The tool accepts the current working directory as input and identifies uncommitted changes
- The tool correctly identifies which files have been modified, added, or deleted
- The tool produces a diff representation of the uncommitted changes
- The tool handles empty working directories (no changes) gracefully with an informative message
- The tool handles repositories with no git history gracefully
- The output clearly indicates which files were included in the evaluation

**Definition of Done**:
- Uncommitted changes are correctly identified and evaluated
- Empty working directories produce a clear message
- The list of evaluated files is visible in the output

---

### Story 2.2: Diff Input — Staged Changes

**Story Title**: Evaluate Staged Changes

**As a** Developer
**I want** to run the quality gate against my staged changes
**So that** I can verify the quality of exactly what I'm about to commit

**Acceptance Criteria**:
- The tool accepts staged (index) changes as input and evaluates only those changes
- The tool correctly distinguishes between staged and unstaged changes in the same file
- The tool evaluates only the staged portion of a partially-staged file
- The tool handles the case where no changes are staged with an informative message

**Definition of Done**:
- Staged changes are correctly identified and evaluated
- Partially staged files evaluate only the staged portion
- Empty staging area produces a clear message

---

### Story 2.3: Diff Input — Committed Changes Against a Base

**Story Title**: Evaluate Committed Changes Against a Base Reference

**As a** Code Reviewer or CI Pipeline Operator
**I want** to run the quality gate against committed changes compared to a base branch or tag
**So that** I can evaluate a feature branch against main, or a specific commit range, from either a local review or CI context

**Acceptance Criteria**:
- The tool accepts a base reference (branch name, tag, or commit hash) and compares the current HEAD against it
- The tool correctly produces the diff between the current state and the specified base
- The tool handles the case where the base reference does not exist with a clear error
- The tool handles the case where the base and current state are identical with an informative message

**Definition of Done**:
- Committed changes against any valid base reference are correctly identified and evaluated
- Invalid base references produce clear error messages
- Identical base and current state produce a clear message

---

### Story 2.4: Diff Input — Arbitrary Diff File

**Story Title**: Evaluate an Arbitrary Diff File

**As a** Developer or CI Pipeline Operator
**I want** to run the quality gate against a diff file from any source
**So that** I can evaluate changes generated by tools other than git, or evaluate historical diffs for analysis

**Acceptance Criteria**:
- The tool accepts a diff file (in unified diff format) as input from a file path or standard input
- The tool correctly parses the diff file and identifies changed files and hunks
- The tool handles malformed diff files with a clear error message
- The tool handles empty diff files with an informative message
- The tool handles diff files with no meaningful code changes (e.g., whitespace-only) with an appropriate response

**Definition of Done**:
- Arbitrary diff files are correctly parsed and evaluated
- Malformed diff files produce clear error messages
- Empty or whitespace-only diffs are handled gracefully

---

### Story 2.5: Diff Filtering

**Story Title**: Filter Irrelevant Files from Evaluation

**As a** Developer
**I want** the quality gate to automatically exclude lock files, generated code, and binary assets from evaluation
**So that** the evaluation focuses only on meaningful code changes and avoids noise and unnecessary cost

**Acceptance Criteria**:
- Lock files (e.g., package-lock.json, yarn.lock, go.sum) are excluded by default
- Generated code files (identified by common patterns or markers) are excluded by default
- Binary asset files are excluded by default
- The exclusion list can be customized in the configuration file
- The exclusion list can be extended with additional patterns in the configuration file
- The output indicates which files were excluded and why
- No files are excluded unless they match a configured pattern

**Definition of Done**:
- Default exclusion patterns correctly filter common non-evaluable files
- Custom exclusion patterns can be added via configuration
- Excluded files are reported in the output

---

### Story 2.6: Code Quality Evaluation

**Story Title**: Evaluate Code Quality of Diff Changes

**As a** Developer
**I want** the quality gate to evaluate the structural quality of my code changes
**So that** I can identify anti-patterns, maintainability issues, and structural problems before merging

**Acceptance Criteria**:
- The tool evaluates code changes for structural problems such as deeply nested logic, overly complex functions, and duplicated code patterns
- The tool evaluates code changes for common anti-patterns such as god objects, feature envy, and long parameter lists
- The tool evaluates code changes for maintainability concerns such as unclear naming, missing documentation on public interfaces, and magic numbers
- The evaluation produces specific, line-level feedback identifying where each issue occurs
- The evaluation handles multi-language diffs by applying language-appropriate criteria
- Code changes with no quality issues receive positive confirmation, not silence

**Definition of Done**:
- Quality evaluation produces specific, line-level feedback for each issue found
- Clean code receives positive confirmation
- Multi-language diffs are evaluated with language-appropriate criteria

---

### Story 2.7: Test Coverage Adequacy Evaluation

**Story Title**: Evaluate Test Coverage Adequacy for Changed Code Paths

**As a** Developer or QA Engineer
**I want** the quality gate to assess whether my changed code paths are adequately tested
**So that** test-free code and code with meaningless tests does not enter the codebase

**Acceptance Criteria**:
- The tool assesses whether new functions and methods introduced in the diff have corresponding test coverage
- The tool assesses whether modified code paths have updated or new tests
- The tool goes beyond numeric coverage to evaluate whether tests actually exercise the changed behavior
- The tool consumes pre-computed coverage data from existing test tooling (does not generate coverage itself)
- The tool identifies specific code paths that lack test coverage
- The tool handles the case where no coverage data is available with a clear indication that coverage assessment is incomplete

**Definition of Done**:
- Coverage assessment identifies specific untested code paths
- Tests are evaluated for meaningfulness, not just presence
- Missing coverage data is communicated clearly to the user

---

### Story 2.8: Deployability Evaluation

**Story Title**: Evaluate Production Deployability of Changes

**As a** Release Manager or DevOps Engineer
**I want** the quality gate to assess whether my changes are ready for production deployment
**So that** I can catch deployment blockers before they reach the merge stage

**Acceptance Criteria**:
- The tool evaluates whether changes introduce hardcoded credentials, secrets, or sensitive configuration values
- The tool evaluates whether changes include migration scripts or configuration updates required for deployment
- The tool evaluates whether changes introduce dependencies that require deployment-time resolution
- The tool evaluates whether changes include proper error handling for production failure modes
- The tool identifies specific deployment blockers with line-level references
- Changes with no deployability concerns receive positive confirmation

**Definition of Done**:
- Deployability evaluation identifies specific blockers with line-level references
- Clean changes receive positive confirmation
- Deployment concerns are clearly categorized and actionable

---

### Story 2.9: Large Diff Handling

**Story Title**: Handle Large Diffs Gracefully

**As a** Developer working on a large refactoring or migration
**I want** the quality gate to handle diffs larger than 500 lines without failing silently or producing unreliable results
**So that** I can still get quality feedback on large changes without having to split them artificially

**Acceptance Criteria**:
- Diffs exceeding 500 lines of changes are evaluated without error
- The tool provides a clear indication when a diff is large and how it is being handled
- Large diffs do not produce evaluation results of significantly lower quality than small diffs
- The tool completes evaluation of large diffs within the CI latency target (60 seconds)
- The tool provides a warning when a diff exceeds a recommended size threshold
- The output for large diffs remains specific and actionable, not vague or overly summarized

**Definition of Done**:
- Large diffs are evaluated successfully within latency targets
- Users are warned about diff size but evaluation proceeds
- Feedback quality is maintained for large diffs

---

## Epic 3: Farley Score Engine

**Goal:** Implement the holistic scoring system that aggregates quality, coverage, and deployability into a single, intuitively understood metric.

**Rationale:** The Farley Score is the central decision metric. It must be consistent, transparent, and trustworthy. Built after core evaluation pillars are in place.

---

### Story 3.1: Farley Score Computation

**Story Title**: Compute the Farley Score from Evaluation Pillars

**As a** Developer or Code Reviewer
**I want** to see a single quality score (1–10) that aggregates code quality, test coverage, and deployability
**So that** I can make a quick go/no-go decision without interpreting multiple separate metrics

**Acceptance Criteria**:
- The tool computes a single numeric score on a 1–10 scale (the "Farley Score")
- The score aggregates results from all three evaluation pillars: code quality, test coverage adequacy, and deployability
- Each pillar's contribution to the final score is visible in the output
- The score is computed consistently given the same input (same diff, same coverage data, same configuration)
- The score respects configurable pillar weights from the configuration file
- A score of 10 represents exemplary quality across all pillars; a score of 1 represents critical failures across all pillars

**Definition of Done**:
- Farley Score is computed and displayed on a 1–10 scale
- Individual pillar contributions are visible
- Score is consistent across repeated evaluations of the same input
- Pillar weights from configuration are respected

---

### Story 3.2: Score Consistency

**Story Title**: Consistent Scoring Across Environments

**As a** Developer who runs checks locally and in CI
**I want** the quality gate to produce the same Farley Score for the same diff whether I run it locally or in CI
**So that** I can trust the local result as a predictor of the CI result

**Acceptance Criteria**:
- The same diff, coverage data, and configuration produce the same Farley Score whether evaluated locally or in CI
- The score does not vary based on operating system, environment variables, or execution context
- Repeated evaluations of the same input produce the same score (deterministic scoring)
- Any factors that could cause score divergence between environments are identified and controlled

**Definition of Done**:
- Same input produces same score across local and CI environments
- Score is deterministic across repeated evaluations
- No environment-dependent factors affect scoring

---

### Story 3.3: Score Breakdown Transparency

**Story Title**: Transparent Score Breakdown

**As a** Developer whose code was rejected
**I want** to see exactly how the Farley Score was calculated, including each pillar's score and the specific issues that lowered it
**So that** I understand exactly what to fix and can challenge the score if I believe it is incorrect

**Acceptance Criteria**:
- The output shows the final Farley Score alongside individual pillar scores
- Each pillar score is accompanied by the specific issues that contributed to it
- The output shows the weight applied to each pillar and how it affected the final score
- The output is presented in a format that is easy to scan and understand
- The breakdown is available in both local and CI contexts

**Definition of Done**:
- Score breakdown shows final score, pillar scores, weights, and contributing issues
- Breakdown is easy to scan and understand
- Available in both local and CI output

---

## Epic 4: Local Advisory Mode

**Goal:** Deliver the developer workstation experience — fast, advisory feedback that helps developers self-correct before pushing.

**Rationale:** Local mode is the "shift-left" mechanism. It must be fast (< 10 seconds), non-blocking, and actionable. Built after core evaluation is solid.

---

### Story 4.1: Local Advisory Output

**Story Title**: Advisory Feedback on Developer Workstation

**As a** Developer
**I want** to run the quality gate locally and receive clear feedback about my code quality without being blocked from committing or pushing
**So that** I can fix issues before pushing and avoid CI rejections

**Acceptance Criteria**:
- The tool runs on the developer's workstation and evaluates the current changes
- When the Farley Score is above the threshold, the tool displays a positive message with the score
- When the Farley Score is below the threshold, the tool displays a warning with the score and specific improvement recommendations
- The tool never prevents the developer from committing or pushing, regardless of score
- The output is human-readable and formatted for terminal display
- The evaluation completes within 10 seconds for typical diffs (under 500 lines)

**Definition of Done**:
- Local evaluation produces clear, terminal-friendly output
- Above-threshold scores display positive feedback
- Below-threshold scores display warnings with specific recommendations
- Developer workflow is never blocked
- Evaluation completes within 10 seconds for typical diffs

---

### Story 4.2: Local Structured Output

**Story Title**: Machine-Readable Output for IDE Integration

**As a** Developer using an IDE or editor plugin
**I want** the quality gate to produce structured output (JSON) that my editor can parse
**So that** I can see quality feedback directly in my editor without switching to a terminal

**Acceptance Criteria**:
- The tool supports a flag or option to produce output in JSON format
- The JSON output includes the Farley Score, pillar scores, and all issue details
- Each issue in the JSON output includes the file path, line number, severity, and description
- The JSON output is valid and parseable by standard JSON parsers
- The JSON output contains the same information as the human-readable output
- The default output format remains human-readable; JSON is opt-in

**Definition of Done**:
- JSON output is valid and contains all evaluation details
- Each issue includes file path, line number, severity, and description
- JSON format is opt-in; default remains human-readable

---

### Story 4.3: Local Latency Expectation

**Story Title**: Near-Instant Local Feedback

**As a** Developer in the middle of a coding session
**I want** the quality gate to return results within seconds when running locally
**So that** I can stay in my flow and not abandon the local check due to waiting

**Acceptance Criteria**:
- Evaluation of diffs under 500 lines completes within 10 seconds when run locally
- The tool provides feedback or progress indication if evaluation is taking longer than expected
- The tool does not consume excessive CPU, memory, or network bandwidth on the developer's machine
- Repeated local evaluations (e.g., after fixing an issue) complete within the same time target
- The tool handles network latency gracefully and does not hang indefinitely

**Definition of Done**:
- Typical diffs evaluate within 10 seconds locally
- Progress indication is shown for longer evaluations
- Resource consumption on developer machine is reasonable
- Network issues are handled gracefully

---

## Epic 5: CI/CD Authoritative Mode

**Goal:** Deliver the pipeline integration that enforces the quality gate — hard rejection below threshold.

**Rationale:** This is the "gatekeeper" behavior. It must be reliable, fast (< 60 seconds), and produce CI-consumable output. Built after local mode is proven.

---

### Story 5.1: CI Hard Rejection

**Story Title**: Automatic Rejection Below Quality Threshold

**As a** DevOps Engineer or Engineering Manager
**I want** the quality gate to automatically reject pull requests that score below the quality threshold
**So that** substandard code cannot be merged into protected branches

**Acceptance Criteria**:
- When run in a CI/CD pipeline, the tool evaluates the pull request diff
- When the Farley Score is at or above the threshold, the tool signals success and allows the merge to proceed
- When the Farley Score is below the threshold, the tool signals failure and blocks the merge
- The tool uses standard exit codes to signal success (exit 0) and failure (non-zero exit)
- The rejection message includes the Farley Score, the threshold, and specific improvement recommendations
- The rejection message is posted as a comment on the pull request for visibility

**Definition of Done**:
- Below-threshold scores block merge via non-zero exit code
- Above-threshold scores allow merge via zero exit code
- Rejection includes score, threshold, and specific recommendations
- Rejection is posted as a PR comment

---

### Story 5.2: CI Step Summary

**Story Title**: CI Pipeline Step Summary Reporting

**As a** Developer viewing CI pipeline results
**I want** to see a clear summary of the quality gate result in the CI pipeline output
**So that** I can understand the result at a glance without digging into logs

**Acceptance Criteria**:
- The tool produces a concise summary suitable for CI pipeline step output
- The summary includes the Farley Score, threshold, pass/fail status, and a brief reason
- The summary is formatted for CI pipeline display (plain text, no ANSI escape codes unless supported)
- The summary is included in the pipeline step's final output for easy viewing
- The summary distinguishes between pass, fail, and error states

**Definition of Done**:
- CI step summary includes score, threshold, status, and brief reason
- Summary is formatted for CI pipeline display
- Pass, fail, and error states are clearly distinguished

---

### Story 5.3: CI Pull Request Comments

**Story Title**: Detailed Feedback as Pull Request Comments

**As a** Developer whose PR was rejected
**I want** to see the quality gate's detailed feedback as comments on my pull request
**So that** I can see exactly what to fix without navigating to separate CI logs

**Acceptance Criteria**:
- When a PR is rejected, the tool posts a comment on the PR with the full evaluation details
- The PR comment includes the Farley Score, pillar breakdown, and all specific issues with line references
- When a PR passes, the tool posts a comment confirming the score and approval
- The comment is formatted for readability in the PR interface (markdown-compatible)
- The tool does not duplicate comments if the same commit is re-evaluated

**Definition of Done**:
- Rejected PRs receive detailed comments with score, breakdown, and issues
- Passed PRs receive confirmation comments
- Comments are markdown-formatted and readable
- Duplicate comments are not posted for the same commit

---

### Story 5.4: Branch Protection Integration

**Story Title**: Protected Branch Enforcement

**As an** Engineering Manager
**I want** the quality gate to enforce quality standards on protected branches (main, develop)
**So that** the quality floor is maintained on critical branches

**Acceptance Criteria**:
- The tool can be configured to enforce quality standards on specified branches
- Merges to protected branches require the quality gate to pass
- The tool can operate in advisory mode on non-protected branches (feature branches)
- The list of protected branches is configurable in the configuration file
- The tool clearly indicates when it is running in enforcement mode versus advisory mode

**Definition of Done**:
- Protected branches enforce quality gate pass requirement
- Non-protected branches operate in advisory mode by default
- Protected branch list is configurable
- Enforcement vs. advisory mode is clearly indicated

---

## Epic 6: Feedback & Reporting

**Goal:** Ensure every evaluation produces specific, actionable feedback that developers can act on immediately.

**Rationale:** A rejection without actionable feedback is just noise. Feedback must be specific, line-level, and prioritized.

---

### Story 6.1: Actionable Rejection Feedback

**Story Title**: Specific, Actionable Improvement Recommendations

**As a** Developer whose code was rejected
**I want** to receive specific, line-level improvement recommendations that tell me exactly what to fix
**So that** I can address the issues quickly without guessing what the problem is

**Acceptance Criteria**:
- Every rejection includes specific recommendations tied to the diff, not generic advice
- Each recommendation references the specific file and line number where the issue occurs
- Each recommendation describes what the issue is and how to fix it in plain language
- Recommendations are prioritized by severity (critical issues listed first)
- The feedback avoids vague language (e.g., "improve code quality") in favor of specific guidance (e.g., "extract the nested loop on lines 45-52 into a separate function")
- The number of recommendations is proportional to the number of issues found (not overwhelming)

**Definition of Done**:
- All recommendations are specific, line-level, and actionable
- Recommendations are prioritized by severity
- Feedback avoids vague language in favor of specific guidance

---

### Story 6.2: Positive Feedback for Clean Code

**Story Title**: Positive Confirmation for High-Quality Changes

**As a** Developer who wrote clean, well-tested code
**I want** to receive positive confirmation when my code passes the quality gate
**So that** I feel reinforced for writing good code and know the gate is working

**Acceptance Criteria**:
- When code passes the quality gate with a high score, the tool displays positive feedback
- The positive feedback includes the Farley Score and a brief acknowledgment of quality
- The positive feedback is encouraging but not excessive
- The tool does not produce false positives (praising code with obvious issues)

**Definition of Done**:
- High-scoring code receives positive, encouraging feedback
- Feedback includes Farley Score
- No false positives (issues are not overlooked)

---

### Story 6.3: Evaluation Evidence

**Story Title**: Raw Evaluation Data for Auditability

**As a** Developer or Engineering Manager
**I want** to see the raw data and evidence the quality gate used to make its judgment
**So that** I can verify the evaluation was fair and challenge it if needed

**Acceptance Criteria**:
- The tool provides access to the raw data used in the evaluation (diff analyzed, coverage data consumed, configuration applied)
- The evidence is available in both local and CI contexts
- The evidence can be exported or saved for later review
- The evidence includes a timestamp and version identifier for reproducibility
- Sensitive information in the evidence is handled appropriately (no credentials or secrets exposed)

**Definition of Done**:
- Raw evaluation data is accessible in both local and CI contexts
- Evidence includes timestamp and version for reproducibility
- Sensitive information is not exposed in evidence

---

## Epic 7: Reliability & Fallback

**Goal:** Ensure the quality gate degrades gracefully when the LLM API is unavailable — it should never permanently block deployments.

**Rationale:** API outages happen. The gate must have a fallback path that preserves pipeline progress while maintaining quality standards as best as possible.

---

### Story 7.1: LLM API Unavailable Fallback

**Story Title**: Graceful Degradation When LLM API Is Unavailable

**As a** DevOps Engineer
**I want** the quality gate to have a fallback path when the LLM API is unavailable
**So that** a temporary API outage does not permanently block all deployments

**Acceptance Criteria**:
- When the LLM API is unavailable, the tool detects the failure and activates a fallback mode
- In fallback mode, the tool performs a best-effort evaluation using deterministic rules (e.g., file type checks, basic pattern matching)
- The fallback evaluation produces a score and feedback, clearly labeled as a fallback result
- The fallback result is treated as advisory (warning) rather than authoritative (hard rejection) in CI mode
- The tool logs the fallback activation for operational visibility
- The tool retries the LLM API before giving up and falling back

**Definition of Done**:
- API unavailability triggers fallback mode with clear indication
- Fallback evaluation produces a score using deterministic rules
- Fallback results are advisory, not authoritative, in CI mode
- Fallback activation is logged for visibility

---

### Story 7.2: Retry Behavior

**Story Title**: Automatic Retry on Transient API Failures

**As a** Developer or CI Pipeline Operator
**I want** the quality gate to automatically retry when the LLM API experiences a transient failure
**So that** temporary network issues do not cause unnecessary evaluation failures

**Acceptance Criteria**:
- The tool automatically retries LLM API calls when a transient error occurs (e.g., timeout, rate limit)
- The number of retries is configurable with a sensible default
- The tool waits between retries with increasing intervals (exponential backoff)
- The tool reports the number of retries attempted before success or final failure
- After all retries are exhausted, the tool activates the fallback path
- The total retry time does not exceed the latency target for the execution context

**Definition of Done**:
- Transient API failures trigger automatic retries with exponential backoff
- Retry count is configurable
- Retries fall back to fallback mode after exhaustion
- Total retry time respects latency targets

---

### Story 7.3: Cost Monitoring

**Story Title**: Evaluation Cost Awareness

**As a** DevOps Engineer or Engineering Manager
**I want** to know the cost of each quality gate evaluation
**So that** I can monitor spending and ensure the tool remains cost-effective

**Acceptance Criteria**:
- The tool reports the estimated cost of each evaluation in the output
- The cost report includes a breakdown by evaluation pillar (quality, coverage, deployability)
- The tool provides a cumulative cost summary when multiple evaluations are run in a session
- The tool warns when a single evaluation's estimated cost exceeds a configurable limit
- The cost reporting works in both local and CI contexts

**Definition of Done**:
- Per-evaluation cost is reported with pillar breakdown
- Cumulative cost is tracked across sessions
- Cost limit warnings are configurable and functional

---

## Epic 8: Operational Visibility

**Goal:** Provide visibility into quality trends, rejection rates, and adoption metrics for engineering leadership.

**Rationale:** Managers need data to assess whether the gate is improving quality, not just blocking merges.

---

### Story 8.1: Quality Trend Reporting

**Story Title**: Code Quality Trend Reports

**As an** Engineering Manager
**I want** to see trends in code quality scores over time
**So that** I can assess whether the quality gate is improving team code quality

**Acceptance Criteria**:
- The tool records evaluation results (score, timestamp, branch, author) for trend analysis
- The tool can produce a summary report of quality scores over a configurable time period
- The report shows average score, score distribution, and trend direction (improving, stable, declining)
- The report can be filtered by branch, author, or time period
- The report is available in both human-readable and machine-readable formats

**Definition of Done**:
- Evaluation results are recorded for trend analysis
- Trend reports show average, distribution, and direction
- Reports are filterable by branch, author, and time period

---

### Story 8.2: Rejection Category Analytics

**Story Title**: Rejection Category Breakdown

**As an** Engineering Manager
**I want** to see which categories of issues are causing the most rejections
**So that** I can target training and process improvements to address the most common quality gaps

**Acceptance Criteria**:
- The tool tracks rejection reasons by category (code quality, test coverage, deployability)
- The tool can produce a report showing the frequency of each rejection category over time
- The report identifies the top rejection categories and their trends
- The report can be filtered by team, branch, or time period
- The report is available in both human-readable and machine-readable formats

**Definition of Done**:
- Rejection reasons are tracked by category
- Reports show frequency and trends of each category
- Reports are filterable by team, branch, and time period

---

### Story 8.3: Local Adoption Tracking

**Story Title**: Local Check Adoption Metrics

**As an** Engineering Manager
**I want** to know whether developers are using the local quality check proactively
**So that** I can assess the shift-left benefit and identify adoption barriers

**Acceptance Criteria**:
- The tool tracks when local evaluations are run (timestamp, user, repository)
- The tool can produce a report showing local check frequency per developer
- The report correlates local check usage with CI pass rates (do developers who check locally have higher CI pass rates?)
- The report respects developer privacy (no code content is tracked, only metadata)
- The adoption tracking is opt-in and configurable

**Definition of Done**:
- Local check usage is tracked as metadata only (no code content)
- Adoption reports show frequency per developer and correlation with CI pass rates
- Tracking is opt-in and configurable

---

## Epic 9: Security & Credential Management

**Goal:** Handle API credentials securely and address data privacy concerns when sending code diffs to external LLM APIs.

**Rationale:** Code diffs may contain proprietary IP. Security and compliance teams need confidence that data is handled appropriately.

---

### Story 9.1: API Credential Management

**Story Title**: Secure API Credential Handling

**As a** Developer or DevOps Engineer
**I want** to configure my LLM API credentials securely without embedding them in code or configuration files
**So that** my credentials are not exposed in version control or shared inadvertently

**Acceptance Criteria**:
- API credentials are configured through environment variables or a secure credential store
- The tool never writes API credentials to disk or to the configuration file
- The tool validates that credentials are configured before attempting evaluation
- Missing credentials produce a clear error message with instructions on how to configure them
- The tool supports per-developer credentials for local use and shared credentials for CI use
- Credentials are not logged or included in error messages

**Definition of Done**:
- Credentials are configured via environment variables or secure store
- Credentials are never written to disk or configuration files
- Missing credentials produce clear setup instructions
- Credentials are not exposed in logs or error messages

---

### Story 9.2: Sensitive Data Pre-Filtering

**Story Title**: Pre-Filter Sensitive Data from Diffs

**As a** Security or Compliance Officer
**I want** the quality gate to filter out sensitive data (credentials, secrets, tokens) from diffs before sending them to the LLM API
**So that** proprietary or sensitive information is not exposed to external services

**Acceptance Criteria**:
- The tool detects common patterns of sensitive data (API keys, passwords, tokens, certificates) in diffs
- Detected sensitive data is redacted before the diff is sent to the LLM API
- The redaction is indicated in the evaluation output so the user knows data was filtered
- The sensitive data patterns can be customized in the configuration file
- The tool does not store or log the redacted sensitive data
- The redaction does not significantly impact evaluation accuracy

**Definition of Done**:
- Common sensitive data patterns are detected and redacted before API calls
- Redaction is indicated in the output
- Custom sensitive data patterns can be configured
- Redacted data is not stored or logged

---

### Story 9.3: Self-Hosted Model Support

**Story Title**: Support for Self-Hosted LLM Models

**As a** Security Officer or Engineering Manager in a regulated environment
**I want** the quality gate to support self-hosted LLM models as an alternative to external APIs
**So that** code diffs never leave our infrastructure

**Acceptance Criteria**:
- The tool supports connecting to a self-hosted LLM model endpoint
- The configuration file allows specifying a self-hosted model endpoint instead of an external API
- The evaluation behavior and output format is identical whether using an external API or self-hosted model
- The tool validates the self-hosted model endpoint before evaluation
- Connection failures to a self-hosted model activate the fallback path

**Definition of Done**:
- Self-hosted model endpoints can be configured
- Evaluation behavior is identical to external API usage
- Self-hosted model failures activate the fallback path

---

## Epic 10: Multi-Language & Repository Support

**Goal:** Handle repositories with multiple programming languages and non-code changes appropriately.

**Rationale:** Real-world repositories are heterogeneous. The gate must handle this without breaking or producing irrelevant feedback.

---

### Story 10.1: Multi-Language Diff Evaluation

**Story Title**: Evaluate Diffs Across Multiple Programming Languages

**As a** Developer working in a polyglot repository
**I want** the quality gate to evaluate code changes across multiple programming languages using language-appropriate criteria
**So that** I get relevant quality feedback regardless of which language I'm working in

**Acceptance Criteria**:
- The tool identifies the programming language of each file in the diff
- The tool applies language-appropriate evaluation criteria for each language
- The tool handles files in languages it does not recognize gracefully (skips with a note)
- The Farley Score aggregates across all languages in the diff
- The output clearly indicates which language criteria were applied to which files

**Definition of Done**:
- Each file is evaluated with language-appropriate criteria
- Unrecognized languages are skipped with a note
- Farley Score aggregates across all languages
- Language-specific criteria are indicated in the output

---

### Story 10.2: Non-Code Change Handling

**Story Title**: Handle Documentation-Only and Configuration Changes

**As a** Developer submitting a documentation update or configuration change
**I want** the quality gate to handle non-code changes appropriately without producing irrelevant feedback
**So that** documentation and configuration PRs are not rejected for reasons that don't apply to them

**Acceptance Criteria**:
- The tool identifies when a diff contains only documentation changes and evaluates accordingly
- The tool identifies when a diff contains only configuration or infrastructure-as-code changes
- Documentation-only changes are evaluated for documentation quality (clarity, completeness) rather than code quality
- Configuration-only changes are evaluated for configuration best practices rather than code quality
- The tool does not flag documentation changes for missing tests or code anti-patterns
- Mixed diffs (code + documentation) evaluate each portion with appropriate criteria

**Definition of Done**:
- Documentation-only diffs are evaluated for documentation quality
- Configuration-only diffs are evaluated for configuration best practices
- No irrelevant feedback is produced for non-code changes
- Mixed diffs apply appropriate criteria to each portion

---

## Epic 11: Edge Cases & Error Handling

**Goal:** Handle edge cases and error scenarios gracefully without crashing or producing misleading results.

**Rationale:** Real-world usage will encounter edge cases. The tool must be robust.

---

### Story 11.1: Empty and Minimal Diffs

**Story Title**: Handle Empty and Minimal Changes

**As a** Developer making a tiny fix (e.g., one-line typo correction)
**I want** the quality gate to handle very small changes without over-evaluating or producing excessive feedback
**So that** trivial changes are not blocked or flagged disproportionately

**Acceptance Criteria**:
- The tool handles diffs with only one or two lines of changes
- The tool does not produce excessive feedback for trivial changes
- The tool evaluates trivial changes fairly and does not penalize them for lacking tests or documentation
- The Farley Score for trivial changes reflects their actual impact, not their size
- The tool clearly indicates when a diff is trivial and how that affected the evaluation

**Definition of Done**:
- Very small diffs are evaluated fairly without excessive feedback
- Trivial changes are not penalized disproportionately
- Trivial diff handling is indicated in the output

---

### Story 11.2: Merge Conflict and Stacked PR Handling

**Story Title**: Evaluate Diffs with Merge Conflicts or Stacked Changes

**As a** Developer working on stacked PRs or resolving merge conflicts
**I want** the quality gate to handle diffs that include merge conflict markers or depend on other unmerged changes
**So that** I can still get quality feedback during complex merge scenarios

**Acceptance Criteria**:
- The tool detects merge conflict markers in the diff and handles them gracefully
- The tool evaluates the developer's changes separately from merge conflict artifacts
- The tool indicates when merge conflict markers are present and how they were handled
- The tool handles stacked PRs where the base includes unmerged changes
- The evaluation remains meaningful even in the presence of merge artifacts

**Definition of Done**:
- Merge conflict markers are detected and handled gracefully
- Developer changes are evaluated separately from merge artifacts
- Presence of merge artifacts is indicated in the output

---

### Story 11.3: Borderline Score Handling

**Story Title**: Handle Borderline Scores Near the Threshold

**As a** Developer whose code scores just below the threshold (e.g., 5.8 out of 6.0)
**I want** to understand whether my code is genuinely below the quality bar or just borderline
**So that** I can decide whether to invest effort in improvements or seek a human review

**Acceptance Criteria**:
- When the Farley Score is within 0.5 points of the threshold, the output indicates the score is borderline
- Borderline scores include a clear message about the proximity to the threshold
- Borderline scores include specific, prioritized recommendations to cross the threshold
- The tool does not treat borderline scores differently from clearly-failing scores in terms of pass/fail decision
- The borderline indication is present in both local and CI contexts

**Definition of Done**:
- Scores within 0.5 of threshold are flagged as borderline
- Borderline scores include prioritized recommendations
- Pass/fail decision is not affected by borderline status
- Borderline indication is present in both contexts

---

## Sequencing Rationale

The stories are sequenced to follow a **risk-reduction-first, value-incremental** approach:

1. **Epic 1 (Foundation)** must come first because without shared configuration, local and CI evaluations diverge immediately, destroying trust. This is the single biggest risk to the system's credibility.

2. **Epic 2 (Core Diff Evaluation)** is the heart of the product. Stories are ordered by input complexity (working directory → staged → committed → arbitrary diff) and evaluation depth (filtering → quality → coverage → deployability → large diffs). Each story delivers a usable evaluation capability.

3. **Epic 3 (Farley Score)** depends on Epic 2's evaluation pillars being in place. The score is the central decision metric and must be trustworthy before building enforcement on top of it.

4. **Epic 4 (Local Advisory)** and **Epic 5 (CI Authoritative)** are parallel tracks that depend on Epic 3. Local mode is built first because it is lower-risk (advisory only) and provides early developer feedback. CI mode is built after local mode is proven.

5. **Epic 6 (Feedback)** runs in parallel with Epics 4–5 because actionable feedback is required for both modes to be useful.

6. **Epic 7 (Reliability)** is built early enough to protect CI pipelines from API outages but after core evaluation is functional.

7. **Epic 8 (Operational Visibility)** is lower priority because it requires evaluation data to accumulate first. It is a "nice to have" for leadership until the core system is proven.

8. **Epic 9 (Security)** is critical for regulated environments but can be addressed after core functionality is proven. Self-hosted model support is a "phase 2" concern for most teams.

9. **Epic 10 (Multi-Language)** and **Epic 11 (Edge Cases)** are polish stories that address real-world complexity after the core single-language, clean-diff case is solid.

---

## Summary

This plan decomposes the Gatekeeper tool into **11 Epics** and **39 User Stories** that collectively deliver:

- **A consistent quality evaluation engine** that evaluates code diffs across three pillars (quality, coverage, deployability) and produces a Farley Score (1–10)
- **Two behavioral modes**: advisory (local, non-blocking) and authoritative (CI, hard rejection)
- **Shared configuration** ensuring identical evaluation rules in both environments
- **Actionable, line-level feedback** for every evaluation outcome
- **Graceful degradation** when the LLM API is unavailable
- **Security-aware credential management** and sensitive data filtering
- **Operational visibility** into quality trends, rejection patterns, and local adoption

The stories follow the **Elephant Carpaccio** technique — each is the thinnest possible vertical slice that delivers end-to-end user value. The sequencing prioritizes **risk reduction** (configuration consistency, scoring reliability) before **enforcement** (CI hard rejection), and **core functionality** before **polish** (multi-language, edge cases).

Each story is **INVEST-compliant**: Independent (can be developed separately), Negotiable (details can be refined), Valuable (delivers clear user value), Estimable (clear scope), Small (single-sprint scope), and Testable (specific acceptance criteria).

---

## Story Inventory

| Epic | # Stories | Priority | Description |
|---|---|---|---|
| 1. Project Foundation & Configuration | 3 | **Critical** | Shared config, validation, presets |
| 2. Core Diff Evaluation | 9 | **Critical** | Diff inputs, filtering, quality/coverage/deployability evaluation |
| 3. Farley Score Engine | 3 | **Critical** | Score computation, consistency, transparency |
| 4. Local Advisory Mode | 3 | **High** | Terminal output, JSON output, latency |
| 5. CI/CD Authoritative Mode | 4 | **High** | Hard rejection, step summary, PR comments, branch protection |
| 6. Feedback & Reporting | 3 | **High** | Actionable feedback, positive feedback, evidence |
| 7. Reliability & Fallback | 3 | **High** | API fallback, retry, cost monitoring |
| 8. Operational Visibility | 3 | **Medium** | Trend reports, rejection analytics, adoption tracking |
| 9. Security & Credentials | 3 | **Medium** | Credential handling, sensitive data filtering, self-hosted models |
| 10. Multi-Language Support | 2 | **Medium** | Polyglot evaluation, non-code changes |
| 11. Edge Cases & Errors | 3 | **Medium** | Empty diffs, merge conflicts, borderline scores |
| **Total** | **39** | | |



